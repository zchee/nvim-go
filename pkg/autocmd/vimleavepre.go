// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/monitoring"
)

func (a *Autocmd) VimLeavePre(pctx context.Context) {
	ctx, span := monitoring.StartSpan(pctx, "VimLeavePre")
	defer span.End()

	logger.FromContext(ctx).Named("VimLeavePre").Debug("canceled")
	<-ctx.Done()
}
