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

	"cloud.google.com/go/errorreporting"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/zchee/nvim-go/pkg/autocmd"
	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/server"
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
		fmt.Printf("%s:\n  version: %s\n", appName, version)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
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

	env, err := config.Process()
	if err != nil {
		logpkg.Fatalf("env.Process: %+v", err)
	}

	var lv zapcore.Level
	if err := lv.UnmarshalText([]byte(env.LogLevel)); err != nil {
		logpkg.Fatalf("failed to parse log level: %s, err: %v", env.LogLevel, err)
	}
	zapLogger, undo := logger.NewRedirectZapLogger(lv)
	defer undo()
	ctx = logger.NewContext(ctx, zapLogger)
	ctx = trace.NewContext(ctx, &trace.Span{}) // add empty span context

	if gcpProjectID := env.GCPProjectID; gcpProjectID != "" {
		// Stackdriver Profiler
		// profCfg := profiler.Config{
		// 	Service:        appName,
		// 	ServiceVersion: tag,
		// 	MutexProfiling: true,
		// 	ProjectID:      gcpProjectID,
		// }
		// if err := profiler.Start(profCfg); err != nil {
		// 	logpkg.Fatalf("failed to start stackdriver profiler: %v", err)
		// }

		// OpenCensus tracing
		sdOpts := stackdriver.Options{
			ProjectID: gcpProjectID,
			OnError: func(err error) {
				zapLogger.Error("stackdriver.Exporter", zap.Error(fmt.Errorf("could not log error: %v", err)))
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
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		})
		view.RegisterExporter(sd)

		ctx, span := trace.StartSpan(ctx, "main") // start root span
		defer span.End()

		// Stackdriver Error Reporting
		errReportCfg := errorreporting.Config{
			ServiceName:    appName,
			ServiceVersion: tag,
			OnError: func(err error) {
				zapLogger.Error("errorreporting", zap.Error(fmt.Errorf("could not log error: %v", err)))
			},
		}
		errClient, err := errorreporting.NewClient(ctx, gcpProjectID, errReportCfg)
		if err != nil {
			logpkg.Fatalf("failed to create errorreporting client: %v", err)
		}
		defer errClient.Close()
		ctx = context.WithValue(ctx, &errorreporting.Client{}, errClient)
	}

	zapLogger.Info("starting "+appName+" server", zap.Object("env", env))

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

	go func() {
		if err := eg.Wait(); err != nil {
			zapLogger.Fatal("eg.Wait", zap.Error(err))
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-ctx.Done():
		zapLogger.Error("ctx.Done()", zap.Error(ctx.Err()))
		return
	case sig := <-sigc:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			zapLogger.Info("catch signal", zap.String("name", sig.String()))
			return
		}
	}
}

func Main(ctx context.Context, p *plugin.Plugin) error {
	ctx, cancel := context.WithCancel(ctx)
	ctx = logger.NewContext(ctx, logger.FromContext(ctx).Named("main"))

	bctxt := buildctxt.NewContext()
	autocmd.Register(ctx, cancel, p, bctxt, command.Register(ctx, p, bctxt))

	// switch to unix socket rpc-connection
	if n, err := server.Dial(ctx); err == nil {
		p.Nvim = n
	}

	return nil
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
