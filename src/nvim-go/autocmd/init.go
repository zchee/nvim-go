// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import "github.com/garyburd/neovim-go/vim/plugin"

func init() {
	autocmd := new(Autocmd)

	autocmd.bufWritePreChan = make(chan error, 2)
	autocmd.bufWritePostChan = make(chan error, 2)

	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmd.bufWritePre)
	plugin.HandleAutocmd("BufWritePost",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmd.bufWritePost)
	plugin.HandleAutocmd("VimEnter",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmdVimEnter)
}
