// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"context"
	"time"

	"github.com/zchee/nvim-go/src/logger"
	"go.uber.org/zap"
)

// Profile measurement of the time it took to any func and output log file.
// Usage: defer nvim.Profile(time.Now(), "func name")
func Profile(ctx context.Context, start time.Time, name string) {
	elapsed := time.Since(start).Seconds()
	logger.FromContext(ctx).WithOptions(zap.AddCallerSkip(2)).Named("profile").Debug("elapsed", zap.Float64(name, elapsed))
}
