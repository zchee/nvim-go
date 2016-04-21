// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/nvim"
)

func init() {
	plugin.HandleCommand("GoByteOffset",
		&plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"},
		commandByteOffset)
}

func commandByteOffset(v *vim.Vim) error {
	p := v.NewPipeline()

	offset, _ := nvim.ByteOffset(p)
	return nvim.Echomsg(v, offset)
}
