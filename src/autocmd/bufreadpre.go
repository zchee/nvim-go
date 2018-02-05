// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"github.com/zchee/nvim-go/src/nvimutil"
)

type bufReadPreEval struct {}

// BufReadPre gets user config variables and assign to global variable when autocmd BufReadPre.
func (a *Autocmd) BufReadPre(eval *bufReadPreEval) {
	defer nvimutil.Profile(a.ctx, time.Now(), "BufReadPre")
}
