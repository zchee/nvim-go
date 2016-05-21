// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"
	"path/filepath"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmdBufWritePre)
}

type bufwritepreEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func autocmdBufWritePre(v *vim.Vim, eval bufwritepreEval) error {
	dir, _ := filepath.Split(eval.File)

	if config.IferrAutosave {
		go commands.Iferr(v, eval.File)
	}

	if config.MetalinterAutosave {
		go commands.Metalinter(v, eval.Cwd)
	}

	if config.FmtAsync {
		go commands.Fmt(v, dir)
	} else {
		return commands.Fmt(v, dir)
	}

	return nil
}
