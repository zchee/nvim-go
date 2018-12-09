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
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/logger"
)

type Server struct {
	*nvim.Nvim
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

		return &Server{Nvim: n}, nil
	}
}

func Dial(pctx context.Context) (*nvim.Nvim, error) {
	const envNvimListenAddress = "NVIM_LISTEN_ADDRESS"
	addr := os.Getenv(envNvimListenAddress) // NVIM_LISTEN_ADDRESS env can get if launched nvim process
	if addr == "" {
		return nil, errors.Errorf("failed get %s", envNvimListenAddress)
	}

	ctx, cancel := context.WithTimeout(pctx, 1*time.Second)
	defer cancel()

	var n *nvim.Nvim
	dialOpts := []nvim.DialOption{
		nvim.DialContext(ctx),
		nvim.DialServe(false),
		nvim.DialLogf(func(format string, a ...interface{}) {
			logger.FromContext(ctx).Info("", zap.Any(format, a))
		}),
	}

	var tempDelay time.Duration // how long to sleep on accept failure
	var err error
	for {
		n, err = nvim.Dial(addr, dialOpts...)
		if err != nil {
			if tempDelay == 0 {
				tempDelay = 5 * time.Millisecond
			} else {
				tempDelay *= 2
			}
			if max := 1 * time.Second; tempDelay > max {
				tempDelay = max
			}
			logger.FromContext(ctx).Error("Dial error", zap.Error(err), zap.Duration("retrying in", tempDelay))
			timer := time.NewTimer(tempDelay)
			select {
			case <-timer.C:
			}
			continue
		}
		tempDelay = 0

		return n, nil
	}
}

func (s *Server) Serve() error {
	return s.Nvim.Serve()
}

func (s *Server) Close() error {
	return s.Nvim.Close()
}
