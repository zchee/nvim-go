// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"nvim-go/nvimutil"
)

// bufEnterEval represents the current buffer number, windows ID and buffer files directory.
type bufEnterEval struct {
	BufNr int    `eval:"bufnr('%')"`
	WinID int    `eval:"win_getid()"`
	Dir   string `eval:"expand('%:p:h')"`
}

// BufEnter gets the current buffer number, windows ID and set context from the directory structure on BufEnter autocmd.
func (a *Autocmd) BufEnter(eval *bufEnterEval) error {
	defer nvimutil.Profile(a.ctx, time.Now(), "BufEnter")

	a.mu.Lock()
	a.buildContext.BufNr = eval.BufNr
	a.buildContext.WinID = eval.WinID
	a.buildContext.Dir = eval.Dir
	a.mu.Unlock()

	if eval.Dir != "" && a.buildContext.PrevDir != eval.Dir {
		a.buildContext.SetContext(eval.Dir)
	}
	return nil
}
