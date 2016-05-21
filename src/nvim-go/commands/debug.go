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
}

func cmdByteOffset(v *vim.Vim) error {
	offset, _ := buffer.ByteOffset(v, 0, 0)
	return nvim.Echomsg(v, offset)
}
