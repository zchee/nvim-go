// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"time"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/nvimutil"
)

type bufReadPreEval struct{}

// BufReadPre gets user config variables and assign to global variable when autocmd BufReadPre.
func (a *Autocmd) BufReadPre(ctx context.Context, eval *bufReadPreEval) {
	defer nvimutil.Profile(ctx, time.Now(), "BufReadPre")
	span := trace.FromContext(ctx)
	span.SetName("BufReadPre")
	defer span.End()
}
