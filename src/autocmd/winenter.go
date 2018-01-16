// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"github.com/zchee/nvim-go/src/nvimutil"
)

// winEnterEval represents the current buffer number, windows ID and buffer files directory.
type winEnterEval struct {
	BufNr int    `eval:"bufnr('%')"`
	WinID int    `eval:"win_getid()"`
	Dir   string `eval:"expand('%:p:h')"`
}

func (a *Autocmd) WinEnter(eval *winEnterEval) error {
	defer nvimutil.Profile(a.ctx, time.Now(), "WinEnter")

	a.getStatus(eval.BufNr, eval.WinID, eval.Dir)
	if eval.Dir != "" && a.buildContext.PrevDir != eval.Dir {
		a.buildContext.SetContext(eval.Dir)
	}

	return nil
}
