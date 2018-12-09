// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"path/filepath"
	"time"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

type bufWritePreEval struct {
	Cwd  string `eval:"getcwd()"`
	File string `eval:"expand('%:p')"`
}

// BufWritePre run the commands on BufWritePre autocmd.
func (a *Autocmd) BufWritePre(ctx context.Context, eval *bufWritePreEval) {
	defer nvimutil.Profile(ctx, time.Now(), "BufWritePre")
	span := trace.FromContext(ctx)
	span.SetName("BufWritePre")
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
	}
}
