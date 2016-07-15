// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"unsafe"

	"nvim-go/nvim"
)

func (c *Commands) cmdBuffers() error {
	bufs, _ := c.v.Buffers()
	var b []string
	for _, buf := range bufs {
		nr, _ := c.v.BufferNumber(buf)
		b = append(b, fmt.Sprintf("%d", nr))
	}
	return nvim.Echomsg(c.v, "Buffers:", b)
}

func (c *Commands) cmdWindows() error {
	wins, _ := c.v.Windows()
	var w []string
	for _, win := range wins {
		w = append(w, win.String())
	}
	return nvim.Echomsg(c.v, "Windows:", w)
}

func (c *Commands) cmdTabpagas() error {
	tabs, _ := c.v.Tabpages()
	var t []string
	for _, tab := range tabs {
		t = append(t, tab.String())
	}
	return nvim.Echomsg(c.v, "Tabpages:", t)
}

func (c *Commands) cmdByteOffset() error {
	b, err := c.v.CurrentBuffer()
	if err != nil {
		return err
	}
	w, err := c.v.CurrentWindow()
	if err != nil {
		return err
	}

	offset, _ := nvim.ByteOffset(c.v, b, w)
	return nvim.Echomsg(c.v, offset)
}

// Stringtoslicebyte convert string to byte slice use unsafe.
func Stringtoslicebyte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
