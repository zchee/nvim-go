// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

// CmdBuildEval struct type for Eval of GoBuild command.
type CmdBuildEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (c *Commands) cmdBuild(bang bool, eval *CmdBuildEval) {
	go func() {
		err := c.Build(bang, eval)

		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.ctxt.Errlist["Build"] = e
			nvimutil.ErrorList(c.Nvim, c.ctxt.Errlist, true)
		}
	}()
}

// Build builds the current buffer's package use compile tool that
// determined from the directory structure.
func (c *Commands) Build(bang bool, eval *CmdBuildEval) interface{} {
	defer nvimutil.Profile(time.Now(), "GoBuild")
	defer c.ctxt.SetContext(eval.Dir)()

	if !bang {
		bang = config.BuildForce
	}

	cmd, err := c.compileCmd(bang, eval.Cwd)
	if err != nil {
		return errors.WithStack(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if buildErr := cmd.Run(); buildErr != nil && buildErr.(*exec.ExitError) != nil {
		errlist, err := nvimutil.ParseError(stderr.Bytes(), eval.Cwd, &c.ctxt.Build)
		if err != nil {
			return errors.WithStack(err)
		}
		return errlist
	}
	delete(c.ctxt.Errlist, "Build")

	return nvimutil.EchoSuccess(c.Nvim, "GoBuild", fmt.Sprintf("compiler: %s", c.ctxt.Build.Tool))
}

func (c *Commands) compileCmd(bang bool, dir string) (*exec.Cmd, error) {
	cmd := exec.Command(c.ctxt.Build.Tool)
	args := []string{"build"}

	if len(config.BuildFlags) > 0 {
		args = append(args, config.BuildFlags...)
	}

	switch c.ctxt.Build.Tool {
	case "go":
		cmd.Dir = dir
		if !bang {
			args = append(args, "-o", os.DevNull)
		}
	case "gb":
		cmd.Dir = c.ctxt.Build.ProjectRoot
	}
	cmd.Args = append(cmd.Args, args...)

	return cmd, nil
}
