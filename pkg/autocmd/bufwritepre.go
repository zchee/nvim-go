// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"path/filepath"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/monitoring"
)

type bufWritePreEval struct {
	Cwd  string `eval:"getcwd()"`
	File string `eval:"expand('%:p')"`
}

// BufWritePre run the commands on BufWritePre autocmd.
func (a *Autocmd) BufWritePre(pctx context.Context, eval *bufWritePreEval) {
	ctx, span := monitoring.StartSpan(pctx, "BufWritePre")
	defer span.End()

	dir := filepath.Dir(eval.File)

	// Iferr need execute before Fmt function because that function calls "noautocmd write"
	// Also do not use goroutine.
	if config.IferrAutosave {
		err := a.cmd.Iferr(ctx, eval.File)
		if err != nil {
			return
		}
	}

	if config.FmtAutosave {
		go func() {
			a.bufWritePreChan <- a.cmd.Fmt(ctx, dir)
		}()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
