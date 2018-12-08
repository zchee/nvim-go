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

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/ocagent"
	"contrib.go.opencensus.io/exporter/stackdriver"
	gops "github.com/google/gops/agent"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	xerrors "golang.org/x/exp/errors"
	xfmt "golang.org/x/exp/errors/fmt"
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
		xfmt.Printf("%s:\n  version: %s\n", appName, version.Version)
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

	env := config.Process()

	var lv zapcore.Level
	if err := lv.UnmarshalText([]byte(env.LogLevel)); err != nil {
		logpkg.Fatalf("failed to parse log level: %s, err: %v", env.LogLevel, err)
	}
	log, undo := logger.NewRedirectZapLogger(lv)
	defer undo()
	ctx = logger.NewContext(ctx, log)

	// Open socket for using gops to get stacktraces of the daemon.
	if err := gops.Listen(gops.Options{ConfigDir: "/tmp/gops", ShutdownCleanup: true}); err != nil {
		logpkg.Fatalf("unable to start gops: %s", err)
	}
	log.Info("starting gops agent")

	if config.HasGCPProjectID() {
		gcpProjectID := env.GCPProjectID
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		})

		// OpenCensus tracing with OpenCensus agent exporter
		oce, err := ocagent.NewExporter(ocagent.WithInsecure(), ocagent.WithServiceName(appName))
		if err != nil {
			msg := xerrors.New("Failed to create the OpenCensus agent exporter")
			err = xfmt.Errorf("%s: %v", msg, err)
			logpkg.Fatal(err)
		}
		defer oce.Stop()
		trace.RegisterExporter(oce)
		log.Info("opencensus", zap.String("trace", "enabled OpenCensus agent exporter"))

		// OpenCensus tracing with Stackdriver exporter
		sdOpts := stackdriver.Options{
			ProjectID: gcpProjectID,
			OnError: func(err error) {
				log.Error("stackdriver.Exporter", zap.Error(xfmt.Errorf("could not log error: %v", err)))
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
		log.Info("opencensus", zap.String("trace", "enabled Stackdriver exporter"))
		view.RegisterExporter(sd)

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

		// Stackdriver Error Reporting
		errReportConf := errorreporting.Config{
			ServiceName:    appName,
			ServiceVersion: version.Tag,
			OnError: func(err error) {
				log.Error("errorreporting", zap.Error(xfmt.Errorf("could not log error: %v", err)))
			},
		}
		errClient, err := errorreporting.NewClient(ctx, gcpProjectID, errReportConf)
		if err != nil {
			logpkg.Fatalf("failed to create errorreporting client: %v", err)
		}
		defer errClient.Close()
		ctx = context.WithValue(ctx, &errorreporting.Client{}, errClient)
		log.Info("stackdriver", zap.String("errorreporting", "enabled Stackdriver errorreporting"))

		var span *trace.Span
		ctx, span = trace.StartSpan(ctx, "main", trace.WithSampler(trace.AlwaysSample())) // start root span
		defer span.End()
	}

	log.Info(xfmt.Sprintf("starting %s server", appName), zap.Object("env", env))
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
			log.Fatal("eg.Wait", zap.Error(err))
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case <-ctx.Done():
		log.Error("ctx.Done()", zap.Error(ctx.Err()))
		return
	case sig := <-sigc:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			log.Info("catch signal", zap.String("name", sig.String()))
			return
		}
	}
}

func Main(ctx context.Context, p *plugin.Plugin) error {
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
}

func Child(ctx context.Context) error {
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
