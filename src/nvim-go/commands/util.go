// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"unsafe"

	"nvim-go/nvim"

	"github.com/neovim-go/vim"
)

func cmdBuffers(v *vim.Vim) error {
	bufs, _ := v.Buffers()
	var b []string
	for _, buf := range bufs {
		nr, _ := v.BufferNumber(buf)
		b = append(b, fmt.Sprintf("%d", nr))
	}
	return nvim.Echomsg(v, "Buffers:", b)
}

func cmdWindows(v *vim.Vim) error {
	wins, _ := v.Windows()
	var w []string
	for _, win := range wins {
		w = append(w, win.String())
	}
	return nvim.Echomsg(v, "Windows:", w)
}

func cmdTabpagas(v *vim.Vim) error {
	tabs, _ := v.Tabpages()
	var t []string
	for _, tab := range tabs {
		t = append(t, tab.String())
	}
	return nvim.Echomsg(v, "Tabpages:", t)
}

func cmdByteOffset(v *vim.Vim) error {
	b, _ := v.CurrentBuffer()
	w, _ := v.CurrentWindow()
	offset, _ := nvim.ByteOffset(v, b, w)
	return nvim.Echomsg(v, offset)
}

// Stringtoslicebyte convert string to byte slice use unsafe.
func Stringtoslicebyte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
