// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufWritePost",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmdBufWritePost)
}

type bufwritepostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func autocmdBufWritePost(v *vim.Vim, eval *bufwritepostEval) {
	if config.BuildAutosave && !config.FmtAsync {
		go commands.Build(v, false, &commands.CmdBuildEval{
			Cwd: eval.Cwd,
			Dir: eval.Dir,
		})
	}

	if config.TestAutosave {
		go commands.Test(v, []string{}, eval.Dir)
	}
}
