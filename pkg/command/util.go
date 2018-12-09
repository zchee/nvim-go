// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

func (c *Command) cmdBuffers(ctx context.Context) error {
	bufs, _ := c.Nvim.Buffers()
	var b []string
	for _, buf := range bufs {
		nr, _ := c.Nvim.BufferNumber(buf)
		b = append(b, fmt.Sprintf("%d", nr))
	}
	return nvimutil.Echomsg(c.Nvim, "Buffers:", b)
}

func (c *Command) cmdWindows(ctx context.Context) error {
	wins, _ := c.Nvim.Windows()
	var w []string
	for _, win := range wins {
		w = append(w, win.String())
	}
	return nvimutil.Echomsg(c.Nvim, "Windows:", w)
}

func (c *Command) cmdTabpagas(ctx context.Context) error {
	tabs, _ := c.Nvim.Tabpages()
	var t []string
	for _, tab := range tabs {
		t = append(t, tab.String())
	}
	return nvimutil.Echomsg(c.Nvim, "Tabpages:", t)
}

func (c *Command) cmdByteOffset(ctx context.Context) error {
	b, err := c.Nvim.CurrentBuffer()
	if err != nil {
		return err
	}
	w, err := c.Nvim.CurrentWindow()
	if err != nil {
		return err
	}

	offset, _ := nvimutil.ByteOffset(c.Nvim, b, w)
	logger.FromContext(ctx).Info("cmdByteOffset", zap.Int("offset", offset))

	return nvimutil.Echomsg(c.Nvim, offset)
}

func (c *Command) cmdNotify(ctx context.Context, args []string) error {
	return nvimutil.Notify(c.Nvim, args...)
}
