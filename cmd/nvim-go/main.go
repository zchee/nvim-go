// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	logpkg "log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/ocagent"
	"contrib.go.opencensus.io/exporter/stackdriver"
	gops "github.com/google/gops/agent"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/errors/fmt"
	"golang.org/x/sync/errgroup"

	"github.com/zchee/nvim-go/pkg/autocmd"
	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimctx"
	"github.com/zchee/nvim-go/pkg/server"
	"github.com/zchee/nvim-go/pkg/version"
)

const (
	appName = "nvim-go"
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
		fmt.Printf("%s:\n  version: %s\n", appName, version.Version)
		return
	}

	ctx, cancel := context.WithCancel(Context())
	defer cancel()

	if *pluginHost != "" {
		os.Unsetenv("NVIM_GO_DEBUG")               // disable zap output
		ctx = logger.NewContext(ctx, zap.NewNop()) // avoid nil panic on logger.FromContext

		fn := func(p *plugin.Plugin) error {
			return func(ctx context.Context, p *plugin.Plugin) error {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				bctxt := buildctxt.NewContext()
				c := command.Register(ctx, p, bctxt)
				autocmd.Register(ctx, cancel, p, bctxt, c)
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
	case err := <-errc:
		if err != nil {
			logpkg.Fatal(err)
		}
		logpkg.Println("all jobs are finished")
	}

	logpkg.Println("done to the shutdown")
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

	// Open socket for using gops to get stacktraces of the daemon.
	if err := gops.Listen(gops.Options{ConfigDir: "/tmp/gops", ShutdownCleanup: true}); err != nil {
		return fmt.Errorf("unable to start gops: %s", err)
	}
	log.Info("starting gops agent")

	if gcpProjectID := env.GCPProjectID; gcpProjectID != "" {
		// OpenCensus tracing with OpenCensus agent exporter
		oce, err := ocagent.NewExporter(ocagent.WithInsecure(), ocagent.WithServiceName(appName))
		if err != nil {
			return fmt.Errorf("Failed to create the OpenCensus agent exporter: %v", err)
		}
		defer func() {
			errs = multierr.Append(errs, oce.Stop())
		}()

		trace.RegisterExporter(oce)
		log.Info("opencensus", zap.String("trace", "enabled OpenCensus agent exporter"))

		// OpenCensus tracing with Stackdriver exporter
		sdOpts := stackdriver.Options{
			ProjectID: gcpProjectID,
			OnError: func(err error) {
				errs = multierr.Append(errs, fmt.Errorf("stackdriver.Exporter: %v", err))
			},
			MetricPrefix: appName,
			Context:      ctx,
		}
		sd, err := stackdriver.NewExporter(sdOpts)
		if err != nil {
			logpkg.Fatalf("failed to create stackdriver exporter: %v", err)
		}
		defer sd.Flush()

		trace.RegisterExporter(sd)
		view.RegisterExporter(sd)
		log.Info("opencensus", zap.String("trace", "enabled Stackdriver exporter"))

		// Stackdriver Profiler
		profConf := profiler.Config{
			Service:        appName,
			ServiceVersion: version.Tag,
			MutexProfiling: true,
			ProjectID:      gcpProjectID,
		}
		if err := profiler.Start(profConf); err != nil {
			logpkg.Fatalf("failed to start stackdriver profiler: %v", err)
		}
		log.Info("stackdriver", zap.String("profiler", "enabled Stackdriver profiler"))

		var span *trace.Span
		ctx, span = trace.StartSpan(ctx, "main", trace.WithSampler(trace.AlwaysSample())) // start root span
		defer span.End()
	}

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		fn := func(p *plugin.Plugin) error {
			return func(ctx context.Context, p *plugin.Plugin) error {
				ctx, cancel := context.WithCancel(ctx)
				log := logger.FromContext(ctx).Named("main")
				ctx = logger.NewContext(ctx, log)

				bctxt := buildctxt.NewContext()
				autocmd.Register(ctx, cancel, p, bctxt, command.Register(ctx, p, bctxt))

				// switch to unix socket rpc-connection
				if n, err := server.Dial(ctx); err == nil {
					p.Nvim = n
				}

				return nil
			}(ctx, p)
		}
		return Plugin(fn)
	})
	eg.Go(func() error {
		return childServer(ctx)
	})

	log.Info(fmt.Sprintf("starting %s server", appName), zap.Object("env", env))
	if err := eg.Wait(); err != nil {
		log.Fatal("occurred error", zap.Error(err))
	}

	return errs
}

func childServer(ctx context.Context) error {
	log := logger.FromContext(ctx).Named("child")
	ctx = logger.NewContext(ctx, log)

	s, err := server.NewServer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create NewServer")
	}
	go s.Serve()

	s.Subscribe(nvimctx.Method)

	select {
	case <-ctx.Done():
		if err := s.Close(); err != nil {
			log.Fatal("Close", zap.Error(err))
		}
		return nil
	}
}
