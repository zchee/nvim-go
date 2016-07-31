// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"

	"nvim-go/config"

	"github.com/neovim-go/vim"
)

type bufWritePreEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (a *Autocmd) cmdBufWritePre(v *vim.Vim, eval *bufWritePreEval) {
	go a.bufWritePre(v, eval)
}

func (a *Autocmd) bufWritePre(v *vim.Vim, eval *bufWritePreEval) {
	dir := filepath.Dir(eval.File)

	// Iferr need execute before Fmt function because that function calls "noautocmd write"
	// Also do not use goroutine.
	if config.IferrAutosave {
		err := a.c.Iferr(eval.File)
		if err != nil {
			return
		}
	}

	if config.FmtAutosave {
		go func() {
			a.bufWritePreChan <- a.c.Fmt(dir)
		}()
	}
}
