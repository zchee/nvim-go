// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("GoByteOffset",
		&plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"},
		cmdByteOffset)
	plugin.HandleCommand("GoBuffers", &plugin.CommandOptions{}, cmdBuffers)
	plugin.HandleCommand("GoWindows", &plugin.CommandOptions{}, cmdWindows)
	plugin.HandleCommand("GoTabpages", &plugin.CommandOptions{}, cmdTabpagas)
}

func cmdBuffers(v *vim.Vim) error {
	bufs, _ := v.Buffers()
	var b []string
	for _, buf := range bufs {
		b = append(b, buf.String())
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
	offset, _ := buffer.ByteOffset(v, b, w)
	return nvim.Echomsg(v, offset)
}
