// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"time"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/terminal"
	"nvim-go/pathutil"

	"github.com/juju/errors"
)

var (
	runTerm *terminal.Terminal
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
		nvim.ErrorWrap(c.v, errors.Annotate(err, "GoRun"))
		return
	}

	cmd := []string{"go", "run", file}
	cmd = append(cmd, lastCmd...)

	go c.Run(cmd, file)
}

// Run runs the go run command for current buffer's packages.
func (c *Commands) Run(cmd []string, file string) error {
	defer profile.Start(time.Now(), "GoRun")

	if runTerm == nil {
		runTerm = terminal.NewTerminal(c.v, "__GO_RUN__", cmd, config.TerminalMode)
	}
	dir, _ := filepath.Split(file)
	rootDir := pathutil.FindVCSRoot(dir)
	runTerm.Dir = rootDir

	if err := runTerm.Run(cmd); err != nil {
		return err
	}

	return nil
}
