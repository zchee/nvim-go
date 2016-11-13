// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"

	"nvim-go/config"

	"github.com/neovim/go-client/nvim"
)

type bufWritePreEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (a *Autocmd) BufWritePre(v *nvim.Nvim, eval *bufWritePreEval) {
	go a.bufWritePre(v, eval)
}

func (a *Autocmd) bufWritePre(v *nvim.Nvim, eval *bufWritePreEval) {
	dir := filepath.Dir(eval.File)

	// Iferr need execute before Fmt function because that function calls "noautocmd write"
	// Also do not use goroutine.
	if config.IferrAutosave {
		err := a.cmds.Iferr(eval.File)
		if err != nil {
			return
		}
	}

	if config.FmtAutosave {
		go func() {
			a.bufWritePreChan <- a.cmds.Fmt(dir)
		}()
	}
}
