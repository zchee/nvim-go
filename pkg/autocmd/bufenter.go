// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"sync"
	"time"

	"go.opencensus.io/trace"
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
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
func (a *Autocmd) BufEnter(eval *bufEnterEval) {
	defer nvimutil.Profile(a.ctx, time.Now(), "BufEnter")
	span := trace.FromContext(a.ctx)
	span.SetName("BufEnter")
	defer span.End()

	configOnce.Do(func() {
		eval.Cfg.Global.ChannelID = a.Nvim.ChannelID()
		config.Get(a.Nvim, eval.Cfg)
		logger.FromContext(a.ctx).Debug("BufEnter", zap.Any("eval.Config", eval.Cfg))
	})

	a.getStatus(eval.BufNr, eval.WinID, eval.Dir)
	if eval.Dir != "" && a.buildContext.PrevDir != eval.Dir {
		a.buildContext.SetContext(eval.Dir)
	}

	// log := logger.FromContext(a.ctx).Named("BufEnter")
	// a.Nvim.RegisterHandler("nvim_buf_lines_event", func(updates ...interface{}) {
	// 	log.Info("nvim_buf_lines_event", zap.Any("updates", updates))
	// })
	// if ok, err := a.Nvim.AttachBuffer(nvim.Buffer(eval.BufNr), true, make(map[string]interface{})); err != nil || !ok {
	// 	log.Error("", zap.Error(errors.Wrap(err, "failed to attach buffer")))
	// }
}
