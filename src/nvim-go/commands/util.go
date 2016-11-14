// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"

	"nvim-go/nvimutil"
)

func (c *Commands) cmdBuffers() error {
	bufs, _ := c.Nvim.Buffers()
	var b []string
	for _, buf := range bufs {
		nr, _ := c.Nvim.BufferNumber(buf)
		b = append(b, fmt.Sprintf("%d", nr))
	}
	return nvimutil.Echomsg(c.Nvim, "Buffers:", b)
}

func (c *Commands) cmdWindows() error {
	wins, _ := c.Nvim.Windows()
	var w []string
	for _, win := range wins {
		w = append(w, win.String())
	}
	return nvimutil.Echomsg(c.Nvim, "Windows:", w)
}

func (c *Commands) cmdTabpagas() error {
	tabs, _ := c.Nvim.Tabpages()
	var t []string
	for _, tab := range tabs {
		t = append(t, tab.String())
	}
	return nvimutil.Echomsg(c.Nvim, "Tabpages:", t)
}

func (c *Commands) cmdByteOffset() error {
	b, err := c.Nvim.CurrentBuffer()
	if err != nil {
		return err
	}
	w, err := c.Nvim.CurrentWindow()
	if err != nil {
		return err
	}

	offset, _ := nvimutil.ByteOffset(c.Nvim, b, w)
	return nvimutil.Echomsg(c.Nvim, offset)
}
