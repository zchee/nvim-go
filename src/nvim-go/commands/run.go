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

	"github.com/neovim-go/vim"
)

var runTerm *terminal.Terminal

// Run runs the go run command for current buffer's packages.
func Run(v *vim.Vim, cmd []string, file string) error {
	defer profile.Start(time.Now(), "GoRun")

	if runTerm == nil {
		runTerm = terminal.NewTerminal(v, "__GO_TEST__", cmd, config.TerminalMode)
	}
	dir, _ := filepath.Split(file)
	rootDir := pathutil.FindVcsRoot(dir)
	runTerm.Dir = rootDir

	if err := runTerm.Run(cmd); err != nil {
		return err
	}

	return nil
}

func cmdRun(v *vim.Vim, args []string, file string) {
	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		cmd = append(cmd, args...)
	}

	go Run(v, cmd, file)
}
