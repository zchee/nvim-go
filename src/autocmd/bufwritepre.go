// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"
	"time"

	"github.com/zchee/nvim-go/src/config"
	"github.com/zchee/nvim-go/src/nvimutil"
)

type bufWritePreEval struct {
	Cwd  string `eval:"getcwd()"`
	File string `eval:"expand('%:p')"`
}

func (a *Autocmd) bufWritePre(eval *bufWritePreEval) {
	go a.BufWritePre(eval)
}

// BufWritePre run the commands on BufWritePre autocmd.
func (a *Autocmd) BufWritePre(eval *bufWritePreEval) {
	defer nvimutil.Profile(a.ctx, time.Now(), "BufWritePre")

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
		go func() {
			a.bufWritePreChan <- a.cmd.Fmt(dir)
		}()
	}
}
