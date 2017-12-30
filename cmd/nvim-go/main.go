// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof" // For pprof debugging.
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/gops/agent"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/zchee/nvim-go/src/autocmd"
	"github.com/zchee/nvim-go/src/buildctx"
	"github.com/zchee/nvim-go/src/command"
	"github.com/zchee/nvim-go/src/logger"
	"go.uber.org/zap"
)

const (
	envDebug = "NVIM_GO_DEBUG"
	envPprof = "NVIM_GO_PPROF"
)

var (
	debug = os.Getenv(envDebug) != ""
	pprof = os.Getenv(envPprof) != ""
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zapLogger := logger.NewZapLogger()
	undo := zap.RedirectStdLog(zapLogger)
	defer undo()
	ctx = logger.NewContext(ctx, zapLogger)

	registerFn := func(p *plugin.Plugin) error {
		buildctxt := buildctx.NewContext()
		c := command.Register(ctx, p, buildctxt)
		autocmd.Register(ctx, p, buildctxt, c)

		if debug {
			// starts the gops agent
			if err := agent.Listen(&agent.Options{NoShutdownCleanup: true}); err != nil {
				return err
			}

			if pprof {
				const addr = "localhost:14715" // (n: 14)vim-(g: 7)(o: 15)
				zapLogger.Debug("start the pprof debugging", zap.String("listen at", addr))

				// enable the report of goroutine blocking events
				runtime.SetBlockProfileRate(1)
				go log.Println(http.ListenAndServe(addr, nil))
			}
		}

		return nil
	}
	go plugin.Main(registerFn)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigc:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			zapLogger.Debug("main", zap.String("interrupted %s signal", sig.String()))
			return
		}
	}
}
