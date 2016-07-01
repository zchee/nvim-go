// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"

	"nvim-go/commands"
	"nvim-go/config"

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

func autocmdBufWritePre(v *vim.Vim, eval *bufwritepreEval) {
	dir, _ := filepath.Split(eval.File)

	if config.FmtAsync {
		go func() {
			commands.Fmt(v, dir)
			v.Command("noautocmd write")
			if config.BuildAutosave {
				commands.Build(v, config.BuildForce, &commands.CmdBuildEval{
					Cwd: eval.Cwd,
					Dir: dir,
				})
			}
		}()
	} else {
		commands.Fmt(v, dir)
	}

	if config.IferrAutosave {
		go commands.Iferr(v, eval.File)
	}

	if config.MetalinterAutosave {
		go commands.Metalinter(v, eval.Cwd)
	}
}
