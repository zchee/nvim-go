// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monitoring

import (
	"context"

	"go.opencensus.io/trace"
)

func StartSpan(pctx context.Context, name string) (context.Context, *trace.Span) {
	return trace.StartSpan(pctx, name)
}
