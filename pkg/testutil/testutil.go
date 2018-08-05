// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"context"

	"github.com/zchee/nvim-go/pkg/logger"
)

func TestContext(ctx context.Context) context.Context {
	return logger.NewContext(ctx, logger.NewZapLogger())
}
