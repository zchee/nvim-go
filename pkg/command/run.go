// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
	"github.com/zchee/nvim-go/pkg/pathutil"
)

var (
	runTerm     *nvimutil.Terminal
	runLastArgs []string
)

func (c *Command) cmdRun(args []string, file string) {
	go func() {
		if err := c.Run(args, file); err != nil {
			nvimutil.ErrorWrap(c.Nvim, err)
		}
	}()
}

func (c *Command) cmdRunLast(file string) {
	if len(runLastArgs) == 0 {
		err := errors.New("not found GoRun last arguments")
		nvimutil.ErrorWrap(c.Nvim, err)
		return
	}

	go func() {
		if err := c.Run(runLastArgs, file); err != nil {
			nvimutil.ErrorWrap(c.Nvim, err)
		}
	}()
}

// Run runs the go run command for current buffer's packages.
func (c *Command) Run(args []string, file string) error {
	defer nvimutil.Profile(c.ctx, time.Now(), "GoRun")

	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		runLastArgs = args
		cmd = append(cmd, args...)
	}

	if runTerm == nil {
		runTerm = nvimutil.NewTerminal(c.Nvim, "__GO_RUN__", cmd, config.TerminalMode)
	}
	runTerm.Dir = pathutil.FindVCSRoot(filepath.Dir(file))

	if err := runTerm.Run(cmd); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
