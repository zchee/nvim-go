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
	plugin.HandleCommand("GoByteOffset", &plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"}, commandByteOffset)
}

func commandByteOffset(v *vim.Vim) error {
	offset, _ := nvim.ByteOffset(v)
	return nvim.Echomsg(v, "%d", offset)
}
