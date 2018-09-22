// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// CmdBuildEval struct type for Eval of GoBuild command.
type CmdBuildEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Command) cmdBuild(args []string, bang bool, eval *CmdBuildEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Build(c.ctx, args, bang, eval)
	}()

	select {
	case <-c.ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Build", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		case nil:
			// nothing to do
		}
	}
}

// Build builds the current buffers package use compile tool that determined
// from the package directory structure.
func (c *Command) Build(ctx context.Context, args []string, bang bool, eval *CmdBuildEval) interface{} {
	defer nvimutil.Profile(ctx, time.Now(), "Build")

	ctx, span := trace.StartSpan(ctx, "Build")
	defer span.End()

	log := logger.FromContext(c.ctx).With(zap.Strings("args", args), zap.Bool("bang", bang), zap.Any("CmdBuildEval", eval))
	if !bang {
		bang = config.BuildForce
	}

	cmd, err := c.compileCmd(ctx, args, bang, eval.Cwd)
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info("", zap.Any("cmd", cmd))

	if buildErr := cmd.Run(); buildErr != nil {
		if err, ok := buildErr.(*exec.ExitError); ok && err != nil {
			errlist, err := nvimutil.ParseError(ctx, stderr.Bytes(), eval.Cwd, &c.buildContext.Build, nil)
			if err != nil {
				span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
				return errors.WithStack(err)
			}
			return errlist
		}
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: buildErr.Error()})
		return errors.WithStack(buildErr)
	}

	return nvimutil.EchoSuccess(c.Nvim, "GoBuild", fmt.Sprintf("compiler: %s", c.buildContext.Build.Tool))
}

// compileCmd returns the *exec.Cmd corresponding to the compile tool.
func (c *Command) compileCmd(ctx context.Context, args []string, bang bool, dir string) (*exec.Cmd, error) {
	ctx, span := trace.StartSpan(ctx, "compileCmd")
	defer span.End()

	bin, err := exec.LookPath(c.buildContext.Build.Tool)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cmd := exec.CommandContext(ctx, bin, "build")
	cmd.Env = os.Environ()

	if len(config.BuildFlags) > 0 {
		args = append(args, config.BuildFlags...)
	}
	switch c.buildContext.Build.Tool {
	case "go":
		cmd.Dir = dir

		// Outputs the binary to DevNull if without bang
		if !bang || !matchSlice("-o", args) {
			args = append(args, "-o", os.DevNull)
		}
		// add "app" suffix to binary name if enable app-engine build
		if config.BuildAppengine {
			cmd.Args[0] += "app"
		}
	case "gb":
		cmd.Dir = c.buildContext.Build.ProjectRoot

		if config.BuildAppengine {
			cmd.Args = append([]string{cmd.Args[0], "gae"}, cmd.Args[1:]...)
			pkgs, err := fs.GbPackages(cmd.Dir)
			if err != nil {
				return nil, err
			}
			for _, pkg := range pkgs {
				// "gb gae build" doesn't compatible "gb build" arg. actually, "goapp build ..."
				cmd.Args = append(cmd.Args, pkg+string(filepath.Separator)+"...")
			}
		}
	}

	args = append(args, "./...")
	cmd.Args = append(cmd.Args, args...)
	logger.FromContext(c.ctx).Debug("compileCmd", zap.Any("cmd", cmd))

	return cmd, nil
}

func matchSlice(s string, ss []string) bool {
	for _, str := range ss {
		if s == str {
			return true
		}
	}
	return false
}
