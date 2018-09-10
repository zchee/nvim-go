// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

func (a *Autocmd) VimLeavePre() {
	defer nvimutil.Profile(a.ctx, time.Now(), "VimLeavePre")

	a.cancel()
	logger.FromContext(a.ctx).Named("VimLeavePre").Debug("canceled")
}
