// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monitoring

import (
	"context"
	"path"

	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/nvimctx"
)

func StartSpan(pctx context.Context, name string) (context.Context, *trace.Span) {
	return trace.StartSpan(pctx, path.Join(nvimctx.PkgName, name))
}
