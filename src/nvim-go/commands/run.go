// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"time"

	"nvim-go/config"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/terminal"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gorun",
		&plugin.CommandOptions{NArgs: "*", Eval: "expand('%:p')"},
		cmdRun)
}

// Run runs the go run command for current buffer's packages.
func Run(v *vim.Vim, cmd []string) error {
	defer profile.Start(time.Now(), "GoRun")
	term := terminal.NewTerminal(v, cmd, config.TerminalMode)

	if err := term.Run(); err != nil {
		return err
	}

	return nil
}

func cmdRun(v *vim.Vim, args []string, file string) {
	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		cmd = append(cmd, args...)
	}

	go Run(v, cmd)
}
