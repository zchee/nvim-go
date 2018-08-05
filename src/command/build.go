// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/src/config"
	"github.com/zchee/nvim-go/src/logger"
	"github.com/zchee/nvim-go/src/nvimutil"
	"github.com/zchee/nvim-go/src/pathutil"
	"go.uber.org/zap"
)

// CmdBuildEval struct type for Eval of GoBuild command.
type CmdBuildEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Command) cmdBuild(args []string, bang bool, eval *CmdBuildEval) {
	go func() {
		c.errs.Delete("Build")

		err := c.Build(args, bang, eval)
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
		}
	}()
}

// Build builds the current buffers package use compile tool that determined
// from the package directory structure.
func (c *Command) Build(args []string, bang bool, eval *CmdBuildEval) interface{} {
	log := logger.FromContext(c.ctx).With(zap.Strings("args", args), zap.Bool("bang", bang), zap.Any("CmdBuildEval", eval))
	if !bang {
		bang = config.BuildForce
	}

	cmd, err := c.compileCmd(args, bang, eval.Cwd)
	if err != nil {
		return errors.WithStack(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Info("", zap.Any("cmd", cmd))

	if buildErr := cmd.Run(); buildErr != nil {
		if err, ok := buildErr.(*exec.ExitError); ok && err != nil {
			errlist, err := nvimutil.ParseError(stderr.Bytes(), eval.Cwd, &c.buildContext.Build, nil)
			if err != nil {
				return errors.WithStack(err)
			}
			return errlist
		}
		return errors.WithStack(buildErr)
	}

	return nvimutil.EchoSuccess(c.Nvim, "GoBuild", fmt.Sprintf("compiler: %s", c.buildContext.Build.Tool))
}

// compileCmd returns the *exec.Cmd corresponding to the compile tool.
func (c *Command) compileCmd(args []string, bang bool, dir string) (*exec.Cmd, error) {
	bin, err := exec.LookPath(c.buildContext.Build.Tool)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cmd := exec.Command(bin, "build")
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
			pkgs, err := pathutil.GbPackages(cmd.Dir)
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
	logger.FromContext(c.ctx).Info("compileCmd", zap.Any("cmd", cmd))

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
