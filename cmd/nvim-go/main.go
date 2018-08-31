// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go: a Go language development plugin for Neovim written in pure Go.
package main

import (
	"context"
	"flag"
	logpkg "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/zchee/nvim-go/pkg/autocmd"
	"github.com/zchee/nvim-go/pkg/buildctx"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/server"
)

var (
	pluginHost  = flag.String("manifest", "", "Write plugin manifest for `host` to stdout")
	vimFilePath = flag.String("location", "", "Manifest is automatically written to `.vim file`")
)

func init() {
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zapLogger, undo := logger.NewRedirectZapLogger()
	defer undo()
	ctx = logger.NewContext(ctx, zapLogger)

	if *pluginHost != "" {
		os.Unsetenv("NVIM_GO_DEBUG")
		fn := func(p *plugin.Plugin) error {
			return func(ctx context.Context, p *plugin.Plugin) error {
				buildctxt := buildctx.NewContext()
				c := command.Register(ctx, p, buildctxt)
				autocmd.Register(ctx, p, buildctxt, c)
				return nil
			}(ctx, p)
		}
		if err := Plugin(fn); err != nil {
			logpkg.Fatal(err)
		}
		return
	}

	eg := new(errgroup.Group)
	eg, ctx = errgroup.WithContext(ctx)
	eg.Go(func() error {
		fn := func(p *plugin.Plugin) error {
			return Main(ctx, p)
		}
		return Plugin(fn)
	})
	eg.Go(func() error {
		return Child(ctx)
	})

	if os.Getenv("NVIM_GO_DEBUG") != "" {
		const addr = ":14715" // (n: 14)vim-(g: 7)(o: 15)
		zapLogger.Debug("start the pprof debugging", zap.String("listen at", addr))

		// enable the report of goroutine blocking events
		runtime.SetBlockProfileRate(1)
		go logpkg.Println(http.ListenAndServe(addr, nil))
	}

	go func() {
		if err := eg.Wait(); err != nil {
			logger.FromContext(ctx).Fatal("eg.Wait", zap.Error(err))
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigc:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			logger.FromContext(ctx).Info("catch signal", zap.String("name", sig.String()))
			cancel() // avoid goroutine leak
			return
		}
	}
}

func Main(ctx context.Context, p *plugin.Plugin) error {
	ctx = logger.NewContext(ctx, logger.FromContext(ctx).Named("main"))

	buildctxt := buildctx.NewContext()
	c := command.Register(ctx, p, buildctxt)
	autocmd.Register(ctx, p, buildctxt, c)

	n, err := dial(ctx)
	if err != nil {
		return err
	}
	p.Nvim = n

	return nil
}

func dial(pctx context.Context) (*nvim.Nvim, error) {
	log := logger.FromContext(pctx).Named("dial")

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
			log.Error("Dial error", zap.Error(err), zap.Duration("retrying in", tempDelay))
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

func Child(ctx context.Context) error {
	log := logger.FromContext(ctx).Named("child")
	ctx = logger.NewContext(ctx, log)

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
