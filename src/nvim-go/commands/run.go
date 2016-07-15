// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"time"

	"nvim-go/config"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/terminal"
	"nvim-go/pathutil"
)

var runTerm *terminal.Terminal

func (c *Commands) cmdRun(args []string, file string) {
	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		cmd = append(cmd, args...)
	}

	go c.Run(cmd, file)
}

// Run runs the go run command for current buffer's packages.
func (c *Commands) Run(cmd []string, file string) error {
	defer profile.Start(time.Now(), "GoRun")

	if runTerm == nil {
		runTerm = terminal.NewTerminal(c.v, "__GO_TEST__", cmd, config.TerminalMode)
	}
	dir, _ := filepath.Split(file)
	rootDir := pathutil.FindVcsRoot(dir)
	runTerm.Dir = rootDir

	if err := runTerm.Run(cmd); err != nil {
		return err
	}

	return nil
}
