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

// BufWritePre run the commands on BufWritePre autocmd.
func (a *Autocmd) BufWritePre(eval *bufWritePreEval) {
	go a.bufWritePre(eval)
}

func (a *Autocmd) bufWritePre(eval *bufWritePreEval) {
	dir := filepath.Dir(eval.File)

	// Iferr need execute before Fmt function because that function calls "noautocmd write"
	// Also do not use goroutine.
	if config.IferrAutosave {
		err := a.cmd.Iferr(eval.File)
		if err != nil {
			return
		}
	}

	if config.FmtAutosave {
		a.mu.Lock()
		go func() {
			defer a.mu.Unlock()

			a.ctx.Errlist = make(map[string][]*nvim.QuickfixError)
			a.bufWritePreChan <- a.cmd.Fmt(dir)
		}()
	}
}
