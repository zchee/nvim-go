// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a Go language development plugin for Neovim written in pure Go.
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
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/src/autocmd"
	"github.com/zchee/nvim-go/src/buildctx"
	"github.com/zchee/nvim-go/src/command"
	"github.com/zchee/nvim-go/src/logger"
	"github.com/zchee/nvim-go/src/server"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zapLogger, undo := logger.NewRedirectZapLogger()
	defer undo()
	ctx = logger.NewContext(ctx, zapLogger)

	var eg = &errgroup.Group{}
	eg, ctx = errgroup.WithContext(ctx)
	eg.Go(func() error {
		fn := func(p *plugin.Plugin) error {
			return Main(ctx, p)
		}
		plugin.Main(fn)
		return nil
	})
	eg.Go(func() error {
		return Child(ctx)
	})
	go func() {
		if err := eg.Wait(); err != nil {
			zapLogger.Fatal("eg.Wait", zap.Error(err))
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigc:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			zapLogger.Info("main", zap.String("interrupted signal", sig.String()))
			cancel() // avoid goroutine leak
			return
		}
	}
}

func Main(ctx context.Context, p *plugin.Plugin) error {
	debug := os.Getenv("NVIM_GO_DEBUG") != ""
	pprof := os.Getenv("NVIM_GO_PPROF") != ""

	log := logger.FromContext(ctx).Named("main")

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

func Child(ctx context.Context) error {
	log := logger.FromContext(ctx).Named("child")

	s, err := server.NewServer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create NewServer")
	}
	go s.Serve()
	defer func() {
		if err := s.Close(); err != nil {
			log.Fatal("Close", zap.Error(err))
		}
	}()

	bufs, err := s.Buffers()
	if err != nil {
		return errors.Wrap(err, "failed to get buffers")
	}
	// Get the names using a single atomic call to Nvim.
	names := make([]string, len(bufs))
	b := s.NewBatch()
	for i, buf := range bufs {
		b.BufferName(buf, &names[i])
	}

	if err := b.Execute(); err != nil {
		return errors.Wrap(err, "failed to execute batch")
	}

	for _, name := range names {
		log.Info("buffer", zap.String("name", name))
	}

	return nil
}
