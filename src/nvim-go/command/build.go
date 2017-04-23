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
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

// CmdBuildEval struct type for Eval of GoBuild command.
type CmdBuildEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Command) cmdBuild(bang bool, eval *CmdBuildEval) {
	go func() {
		err := c.Build(bang, eval)

		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.ctx.Errlist["Build"] = e
			nvimutil.ErrorList(c.Nvim, c.ctx.Errlist, true)
		}
	}()
}

// Build builds the current buffers package use compile tool that determined
// from the package directory structure.
func (c *Command) Build(bang bool, eval *CmdBuildEval) interface{} {
	defer nvimutil.Profile(time.Now(), "GoBuild")

	if !bang {
		bang = config.BuildForce
	}

	cmd, err := c.compileCmd(bang, filepath.Dir(eval.File))
	if err != nil {
		return errors.WithStack(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if buildErr := cmd.Run(); buildErr != nil {
		if buildErr.(*exec.ExitError) != nil {
			errlist, err := nvimutil.ParseError(stderr.Bytes(), eval.Cwd, &c.ctx.Build, nil)
			if err != nil {
				return errors.WithStack(err)
			}
			return errlist
		}
		return errors.WithStack(buildErr)
	}

	// Build succeeded, clean up the Errlist
	delete(c.ctx.Errlist, "Build")

	return nvimutil.EchoSuccess(c.Nvim, "GoBuild", fmt.Sprintf("compiler: %s", c.ctx.Build.Tool))
}

// compileCmd returns the *exec.Cmd corresponding to the compile tool.
func (c *Command) compileCmd(bang bool, dir string) (*exec.Cmd, error) {
	bin, err := exec.LookPath(c.ctx.Build.Tool)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	args := []string{}
	if len(config.BuildFlags) > 0 {
		args = append(args, config.BuildFlags...)
	}

	cmd := exec.Command(bin, "build")
	cmd.Dir = dir

	switch c.ctx.Build.Tool {
	case "go":
		// Outputs the binary to DevNull if without bang
		if !bang && !isTest {
			args = append(args, "-o", os.DevNull)
		}
	case "gb":
		if !isTest {
			cmd.Dir = c.ctx.Build.ProjectRoot
		}
	}

	cmd.Args = append(cmd.Args, args...)

	return cmd, nil
}
