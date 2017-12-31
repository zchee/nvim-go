// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	"context"
	logpkg "log"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/zchee/nvim-go/src/server"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	debug := os.Getenv("NVIM_GO_DEBUG") != ""
	pprof := os.Getenv("NVIM_GO_PPROF") != ""

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zapLogger, undo := logger.NewRedirectZapLogger()
	defer undo()
	ctx = logger.NewContext(ctx, zapLogger)

	registerFn := func(p *plugin.Plugin) error {
		log := logger.FromContext(ctx)

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
				log.Debug("start the pprof debugging", zap.String("listen at", addr))

				// enable the report of goroutine blocking events
				runtime.SetBlockProfileRate(1)
				go logpkg.Println(http.ListenAndServe(addr, nil))
			}
		}

		return nil
	}

	childFn := func(ctx context.Context) {
		log := logger.FromContext(ctx)

		cs, err := server.NewServer(ctx)
		if err != nil {
			log.Error("", zap.Error(err))
		}
		defer cs.Close()

		bufs, err := cs.Nvim.Buffers()
		if err != nil {
			log.Error("", zap.Error(err))
		}

		// Get the names using a single atomic call to Nvim.
		names := make([]string, len(bufs))
		b := cs.Nvim.NewBatch()
		for i, buf := range bufs {
			b.BufferName(buf, &names[i])
		}
		if err := b.Execute(); err != nil {
			log.Error("", zap.Error(err))
		}
		for _, name := range names {
			log.Info("", zap.String("name", name))
		}
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		plugin.Main(registerFn)
		return nil
	})
	eg.Go(func() error {
		childFn(ctx)
		return nil
	})
	if err := eg.Wait(); err != nil {
		zapLogger.Fatal("eg.Wait", zap.Error(err))
	}

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
