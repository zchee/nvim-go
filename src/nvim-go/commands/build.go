// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gobuild", &plugin.CommandOptions{Bang: true, Eval: "[getcwd(), expand('%:p:h')]"}, cmdBuild)
}

// CmdBuildEval struct type for Eval of GoBuild command.
type CmdBuildEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func cmdBuild(v *vim.Vim, bang bool, eval CmdBuildEval) {
	go Build(v, bang, eval)
}

// Build building the current buffer's package use compile tool of determined
// from the directory structure.
func Build(v *vim.Vim, bang bool, eval CmdBuildEval) error {
	defer profile.Start(time.Now(), "GoBuild")
	ctxt := new(context.Build)
	defer ctxt.SetContext(eval.Dir)()

	cmd, err := compileCmd(ctxt, bang, eval)
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		return nvim.EchoSuccess(v, "GoBuild", "")
	}

	if _, ok := err.(*exec.ExitError); ok {
		w, err := v.CurrentWindow()
		if err != nil {
			return err
		}

		loclist, err := quickfix.ParseError(stderr.Bytes(), eval.Cwd, ctxt)
		if err != nil {
			return err
		}
		if err := quickfix.SetLoclist(v, loclist); err != nil {
			return err
		}

		return quickfix.OpenLoclist(v, w, loclist, true)
	}

	return err
}

func compileCmd(ctxt *context.Build, bang bool, eval CmdBuildEval) (*exec.Cmd, error) {
	var (
		compiler = ctxt.Tool
		args     = []string{"build"}
		buildDir string
	)
	if compiler == "go" {
		buildDir = eval.Dir
		if !bang {
			tmpfile, err := ioutil.TempFile(os.TempDir(), "nvim-go")
			if err != nil {
				return nil, err
			}
			defer os.Remove(tmpfile.Name())
			args = append(args, "-o", tmpfile.Name())
		}
	} else if compiler == "gb" {
		buildDir = ctxt.ProjectDir
	}

	cmd := exec.Command(compiler, args...)
	cmd.Dir = buildDir

	return cmd, nil
}
