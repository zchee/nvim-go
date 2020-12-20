// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	logpkg "log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/neovim/go-client/nvim/plugin"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/zchee/nvim-go/pkg/autocmd"
	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nctx"
	"github.com/zchee/nvim-go/pkg/server"
	"github.com/zchee/nvim-go/pkg/version"
)

// flags
var (
	fVersion    = flag.Bool("version", false, "Show the version information.")
	pluginHost  = flag.String("manifest", "", "Write plugin manifest for `host` to stdout")
	vimFilePath = flag.String("location", "", "Manifest is automatically written to `.vim file`")
)

func init() {
	flag.Parse()
	logpkg.SetPrefix("nvim-go: ")
}

func main() {
	if *fVersion {
		fmt.Printf("%s:\n  version: %s\n", nctx.AppName, version.Version)
		return
	}

	ctx, cancel := context.WithCancel(Context())
	defer cancel()

	if *pluginHost != "" {
		os.Unsetenv("NVIM_GO_DEBUG")               // disable zap output
		ctx = logger.NewContext(ctx, zap.NewNop()) // avoid nil panic on logger.FromContext

		fn := func(p *plugin.Plugin) error {
			return func(ctx context.Context, p *plugin.Plugin) error {
				bctxt := buildctxt.NewContext()
				c := command.Register(ctx, p, bctxt)
				autocmd.Register(ctx, p, bctxt, c)
				return nil
			}(ctx, p)
		}
		if err := Plugin(fn); err != nil {
			logpkg.Fatal(err)
		}
		return
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sighupFn := func() {}
	sigintFn := func() {
		logpkg.Println("Start shutdown gracefully")
		cancel()
	}
	go signalHandler(sigc, sighupFn, sigintFn)

	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		errc <- startServer(ctx)
	}()

	select {
	case <-ctx.Done():
	case err := <-errc:
		if err != nil {
			logpkg.Fatal(err)
		}
	}
	logpkg.Println("shutdown nvim-go server")
}

func signalHandler(ch <-chan os.Signal, sighupFn, sigintFn func()) {
	for {
		select {
		case sig := <-ch:
			switch sig {
			case syscall.SIGHUP:
				logpkg.Printf("catch signal %s", sig)
				sighupFn()
			case syscall.SIGINT, syscall.SIGTERM:
				logpkg.Printf("catch signal %s", sig)
				sigintFn()
			}
		}
	}
}

func startServer(ctx context.Context) (errs error) {
	env := config.Process()

	var lv zapcore.Level
	if err := lv.UnmarshalText([]byte(env.LogLevel)); err != nil {
		return fmt.Errorf("failed to parse log level: %s, err: %v", env.LogLevel, err)
	}
	log, undo := logger.NewRedirectZapLogger(lv)
	defer undo()
	ctx = logger.NewContext(ctx, log)

	fn := func(p *plugin.Plugin) error {
		return func(ctx context.Context, p *plugin.Plugin) error {
			log := logger.FromContext(ctx).Named("main")
			ctx = logger.NewContext(ctx, log)

			bctxt := buildctxt.NewContext()
			cmd := command.Register(ctx, p, bctxt)
			autocmd.Register(ctx, p, bctxt, cmd)

			// switch to unix socket rpc-connection
			if n, err := server.Dial(ctx); err == nil {
				p.Nvim = n
			}

			return nil
		}(ctx, p)
	}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return Plugin(fn)
	})
	eg.Go(func() error {
		return subscribeServer(ctx)
	})

	log.Info(fmt.Sprintf("starting %s server", nctx.AppName), zap.Object("env", env))
	if err := eg.Wait(); err != nil {
		log.Fatal("occurred error", zap.Error(err))
	}

	return errs
}

func subscribeServer(ctx context.Context) error {
	log := logger.FromContext(ctx).Named("subscribeServer")
	ctx = logger.NewContext(ctx, log)

	s, err := server.NewServer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create NewServer")
	}
	go s.Serve()

	nctx.RegisterBufLinesEvent(s.Nvim, func(linesEvent ...interface{}) {
		log.Debug(fmt.Sprintf("handles %s", nctx.EventBufLines), zap.Any("linesEvent", linesEvent))
	})
	nctx.RegisterBufChangedtickEvent(s.Nvim, func(changedtickEvent ...interface{}) {
		log.Debug(fmt.Sprintf("handles %s", nctx.EventBufChangedtick), zap.Any("changedtickEvent", changedtickEvent))
	})

	buf := nctx.RegisterBufAttachEvent(s.Nvim, func(attach_event ...interface{}) {
		log.Debug(fmt.Sprintf("handles %s", nctx.EventBufAttach), zap.Any("attach_event", attach_event))
	})
	nctx.RegisterBufDetachEvent(s.Nvim, func(detach_event ...interface{}) {
		log.Debug(fmt.Sprintf("handles %s", nctx.EventBufDetach), zap.Any("detach_event", detach_event))
	})

	select {
	case <-ctx.Done():
		log.Info("Close server")

		if err := s.Close(); err != nil {
			if _, dbErr := s.Nvim.DetachBuffer(buf); dbErr != nil {
				err = multierr.Append(err, dbErr)
			}
			log.Fatal("s.Close", zap.Error(err))
		}
		return nil
	}
}
