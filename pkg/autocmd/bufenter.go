// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/monitoring"
)

// bufEnterEval represents the current buffer number, windows ID and buffer files directory.
type bufEnterEval struct {
	BufNr int    `eval:"bufnr('%')"`
	WinID int    `eval:"win_getid()"`
	Dir   string `eval:"expand('%:p:h')"`

	Cfg *config.Config
}

var configOnce sync.Once

// BufEnter gets the current buffer number, windows ID and set context from the directory structure on BufEnter autocmd.
func (a *Autocmd) BufEnter(pctx context.Context, eval *bufEnterEval) {
	ctx, span := monitoring.StartSpan(pctx, "BufEnter")
	defer span.End()

	configOnce.Do(func() {
		eval.Cfg.Global.ChannelID = a.Nvim.ChannelID()
		config.Get(a.Nvim, eval.Cfg)
		logger.FromContext(ctx).Debug("BufEnter", zap.Any("eval.Config", eval.Cfg))
	})

	a.getStatus(ctx, eval.BufNr, eval.WinID, eval.Dir)
	if eval.Dir != "" && a.buildContext.PrevDir != eval.Dir {
		a.buildContext.SetContext(eval.Dir)
	}
}
