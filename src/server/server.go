// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"os"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/src/logger"
	"go.uber.org/zap"
)

type Server struct {
	Nvim *nvim.Nvim
}

func NewServer(ctx context.Context) (*Server, error) {
	const envNvimListenAddress = "NVIM_LISTEN_ADDRESS"

	addr := os.Getenv(envNvimListenAddress)
	if addr == "" {
		return nil, errors.Errorf("%s not set", envNvimListenAddress)
	}

	zapLogf := func(format string, a ...interface{}) {
		logger.FromContext(ctx).Named("server").Info("", zap.Any(format, a))
	}
	n, err := nvim.Dial(addr, nvim.DialContext(ctx), nvim.DialLogf(zapLogf))
	if err != nil {
		return nil, err
	}

	return &Server{
		Nvim: n,
	}, nil
}

func (s *Server) Close() error {
	return s.Nvim.Close()
}
