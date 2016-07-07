// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"

	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
)

type bufWritePreEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (a *AutocmdContext) autocmdBufWritePre(v *vim.Vim, eval *bufWritePreEval) {
	dir := filepath.Dir(eval.File)

	if config.IferrAutosave {
		err := commands.Iferr(v, eval.File)
		a.send(a.bufWritePreChan, err)
	}
	err := commands.Fmt(v, dir)
	a.send(a.bufWritePreChan, err)
}
