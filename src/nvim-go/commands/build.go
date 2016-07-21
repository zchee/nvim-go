// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/juju/errors"
	"github.com/neovim-go/vim"
)

const pkgBuild = "GoBuild"

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
			nvim.ErrorWrap(c.v, e)
		case []*vim.QuickfixError:
			c.ctxt.Errlist["Build"] = e
			quickfix.ErrorList(c.v, c.ctxt.Errlist, true)
		}
	}()
}

// Build builds the current buffer's package use compile tool that
// determined from the directory structure.
func (c *Commands) Build(bang bool, eval *CmdBuildEval) interface{} {
	defer profile.Start(time.Now(), pkgBuild)
	defer c.ctxt.Build.SetContext(eval.Dir)()

	if !bang {
		bang = config.BuildForce
	}

	cmd, err := c.compileCmd(bang, eval.Cwd)
	if err != nil {
		return errors.Annotate(err, pkgBuild)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if buildErr := cmd.Run(); buildErr != nil && buildErr.(*exec.ExitError) != nil {
		errlist, err := quickfix.ParseError(stderr.Bytes(), eval.Cwd, &c.ctxt.Build)
		if err != nil {
			return errors.Annotate(err, pkgBuild)
		}
		return errlist
	}
	delete(c.ctxt.Errlist, "Build")

	return nvim.EchoSuccess(c.v, pkgBuild, fmt.Sprintf("compiler: %s", c.ctxt.Build.Tool))
}

func (c *Commands) compileCmd(bang bool, dir string) (*exec.Cmd, error) {
	cmd := exec.Command(c.ctxt.Build.Tool)
	args := []string{"build"}

	if len(config.BuildArgs) > 0 {
		args = append(args, config.BuildArgs...)
	}

	switch c.ctxt.Build.Tool {
	case "go":
		cmd.Dir = dir
		if !bang {
			tmpfile, err := ioutil.TempFile(os.TempDir(), "nvim-go")
			if err != nil {
				return nil, err
			}
			defer os.Remove(tmpfile.Name())
			args = append(args, "-o", tmpfile.Name())
		}
	case "gb":
		cmd.Dir = c.ctxt.Build.ProjectRoot
	}
	cmd.Args = append(cmd.Args, args...)

	return cmd, nil
}
