// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/logger"
)

func TestContext(tb testing.TB, ctx context.Context) context.Context {
	tb.Helper()

	return logger.NewContext(ctx, zap.NewNop())
}
