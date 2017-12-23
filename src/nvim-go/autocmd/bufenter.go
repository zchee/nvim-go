// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

// bufEnterEval represents the current buffer number, windows ID and buffer files directory.
type bufEnterEval struct {
	BufNr int    `eval:"bufnr('%')"`
	WinID int    `eval:"win_getid()"`
	Dir   string `eval:"expand('%:p:h')"`
}

// BufEnter gets the current buffer number, windows ID and set context from the directory structure on BufEnter autocmd.
func (a *Autocmd) BufEnter(eval *bufEnterEval) error {
	a.mu.Lock()
	a.buildctxt.BufNr = eval.BufNr
	a.buildctxt.WinID = eval.WinID
	a.buildctxt.Dir = eval.Dir
	a.mu.Unlock()

	a.buildctxt.SetContext(eval.Dir)
	return nil
}
