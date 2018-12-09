// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"time"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

func (a *Autocmd) VimLeavePre(ctx context.Context) {
	defer nvimutil.Profile(ctx, time.Now(), "VimLeavePre")
	span := trace.FromContext(ctx)
	span.SetName("VimLeavePre")
	defer span.End()

	logger.FromContext(ctx).Named("VimLeavePre").Debug("canceled")
	<-ctx.Done()
}
