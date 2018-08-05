// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

func (a *Autocmd) getStatus(bufnr, winID int, dir string) {
	a.mu.Lock()
	a.buildContext.BufNr = bufnr
	a.buildContext.WinID = winID
	a.buildContext.Dir = dir
	a.mu.Unlock()
}
