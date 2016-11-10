// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	"github.com/pkg/errors"
)

var (
	runTerm *nvimutil.Terminal
	lastCmd []string
)

func (c *Commands) cmdRun(args []string, file string) {
	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		lastCmd = args
		cmd = append(cmd, args...)
	}

	go c.Run(cmd, file)
}

func (c *Commands) cmdRunLast(file string) {
	if len(lastCmd) == 0 {
		err := errors.New("not found GoRun last arguments")
		nvimutil.ErrorWrap(c.v, err)
		return
	}

	cmd := []string{"go", "run", file}
	cmd = append(cmd, lastCmd...)

	go c.Run(cmd, file)
}

// Run runs the go run command for current buffer's packages.
func (c *Commands) Run(cmd []string, file string) error {
	defer nvimutil.Profile(time.Now(), "GoRun")

	if runTerm == nil {
		runTerm = nvimutil.NewTerminal(c.v, "__GO_RUN__", cmd, config.TerminalMode)
	}
	dir, _ := filepath.Split(file)
	rootDir := pathutil.FindVCSRoot(dir)
	runTerm.Dir = rootDir

	if err := runTerm.Run(cmd); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
