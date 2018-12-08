// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// VimEnter gets user config variables and assign to global variable when autocmd VimEnter.
func (a *Autocmd) VimEnter(cfg *config.Config) {
	defer nvimutil.Profile(a.ctx, time.Now(), "VimEnter")
	span := trace.FromContext(a.ctx)
	span.SetName("VimEnter")
	defer span.End()
}
