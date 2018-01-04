// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"os"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/src/logger"
	"go.uber.org/zap"
)

type Server struct {
	*nvim.Nvim
	errc chan error
}

func NewServer(pctx context.Context) (*Server, error) {
	log := logger.FromContext(pctx).Named("server")

	const envNvimListenAddress = "NVIM_LISTEN_ADDRESS"
	addr := os.Getenv(envNvimListenAddress)
	if addr == "" {
		return nil, errors.Errorf("%s not set", envNvimListenAddress)
	}

	zapLogf := func(format string, a ...interface{}) {
		log.Info("", zap.Any(format, a))
	}

	ctx, cancel := context.WithTimeout(pctx, 1*time.Second)
	defer cancel()

	var n *nvim.Nvim
	var tempDelay time.Duration
	for {
		var err error
		n, err = nvim.Dial(addr, nvim.DialContext(ctx), nvim.DialServe(false), nvim.DialLogf(zapLogf))
		if err != nil {
			if tempDelay == 0 {
				tempDelay = 5 * time.Millisecond
			} else {
				tempDelay *= 2
			}
			if max := 1 * time.Second; tempDelay > max {
				tempDelay = max
			}
			log.Info("Dial error", zap.Error(err), zap.Duration("retrying in", tempDelay))
			timer := time.NewTimer(tempDelay)
			select {
			case <-timer.C:
			}
			continue
		}
		tempDelay = 0

		return &Server{
			Nvim: n,
			errc: make(chan error, 1),
		}, nil
	}
}

func (s *Server) Serve() {
	s.errc <- s.Nvim.Serve()
}

func (s *Server) Close() error {
	err := s.Nvim.Close()

	var errServe error
	select {
	case errServe = <-s.errc:
	case <-time.After(10 * time.Second):
		errServe = errors.New("nvim: Serve did not exit")
	}
	if err == nil {
		err = errServe
	}

	return err
}
