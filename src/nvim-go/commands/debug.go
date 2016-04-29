// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("GoByteOffset",
		&plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"},
		cmdByteOffset)
}

func cmdByteOffset(v *vim.Vim) error {
	p := v.NewPipeline()

	offset, _ := nvim.ByteOffset(p)
	return nvim.Echomsg(v, offset)
}
