// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nctx

import (
	"context"
	"fmt"

	"github.com/neovim/go-client/nvim"
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/logger"
)

const (
	EventBufLines       = "nvim_buf_lines_event"
	EventBufChangedtick = "nvim_buf_changedtick_event"
)

func RegisterBufLinesEvent(ctx context.Context, n *nvim.Nvim) {
	n.RegisterHandler(EventBufLines, func(linesEvent ...interface{}) {
		logger.FromContext(ctx).Debug(fmt.Sprintf("handles %s", EventBufLines), zap.Any("linesEvent", linesEvent))
	})
}

func RegisterBufChangedtickEvent(ctx context.Context, n *nvim.Nvim) {
	n.RegisterHandler(EventBufChangedtick, func(changedtickEvent ...interface{}) {
		logger.FromContext(ctx).Debug(fmt.Sprintf("handles %s", EventBufChangedtick), zap.Any("changedtickEvent", changedtickEvent))
	})
}
