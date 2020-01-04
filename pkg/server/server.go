// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nctx"
)

type Server struct {
	Nvim *nvim.Nvim
}

func NewServer(ctx context.Context) (*Server, error) {
	n, err := Dial(ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		Nvim: n,
	}, nil
}

var DefaultDialer = func(ctx context.Context, network, address string) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext(ctx, network, address)
}

var DefaultLogFunc = func(ctx context.Context) func(string, ...interface{}) {
	return func(format string, a ...interface{}) {
		logger.FromContext(ctx).Info(fmt.Sprintf(format, a))
	}
}

func Dial(ctx context.Context) (n *nvim.Nvim, err error) {
	log := logger.FromContext(ctx).Named("server")
	ctx = logger.NewContext(ctx, log)

	addr := os.Getenv(nctx.ListenAddress) // NVIM_LISTEN_ADDRESS environment can get if already launch the nvim process
	if addr == "" {
		return nil, errors.Errorf("failed get %s", nctx.ListenAddress)
	}

	dialOpts := []nvim.DialOption{
		nvim.DialContext(ctx),
		nvim.DialServe(false),
		nvim.DialNetDial(DefaultDialer),
		nvim.DialLogf(DefaultLogFunc(ctx)),
	}

	var tmpDelay time.Duration // how long to sleep on accept failure
	for {
		n, err = nvim.Dial(addr, dialOpts...)
		if err != nil {
			log.Error("Dial error", zap.Error(err), zap.Duration("retrying", tmpDelay))

			if tmpDelay == 0 {
				tmpDelay = 5 * time.Millisecond
			} else {
				tmpDelay *= 2
			}
			if max := 1 * time.Second; tmpDelay > max {
				tmpDelay = max
			}
			timer := time.NewTimer(tmpDelay)
			select {
			case <-timer.C:
			}
			continue
		}
		tmpDelay = 0

		return n, nil
	}
}

func (s *Server) Serve() error {
	return s.Nvim.Serve()
}

func (s *Server) Close() error {
	return s.Nvim.Close()
}

type Subscribe struct {
	Server   *Server
	EventMap EventMap
}

type EventMap map[string]func(...interface{})

func NewSubscriber(ctx context.Context, eventmap EventMap) (*Subscribe, error) {
	s, err := NewServer(ctx)
	if err != nil {
		return nil, err
	}

	sb := &Subscribe{
		Server:   s,
		EventMap: eventmap,
	}

	for ev, fn := range sb.EventMap {
		sb.Server.Nvim.RegisterHandler(ev, fn)
		sb.Server.Nvim.Subscribe(ev)
	}

	return sb, nil
}
