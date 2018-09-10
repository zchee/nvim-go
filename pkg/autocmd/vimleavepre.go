// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

func (a *Autocmd) VimLeavePre() {
	defer nvimutil.Profile(a.ctx, time.Now(), "VimLeavePre")

	span := new(trace.Span)
	a.ctx, span = trace.StartSpan(a.ctx, "VimLeavePre")
	defer span.End()

	a.cancel()
	logger.FromContext(a.ctx).Named("VimLeavePre").Debug("canceled")
}
