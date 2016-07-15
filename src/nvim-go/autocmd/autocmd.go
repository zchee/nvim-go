// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"sync"

	"nvim-go/commands"
	"nvim-go/context"

	"github.com/neovim-go/vim"
	"github.com/neovim-go/vim/plugin"
)

// Autocmd represents a autocmd context.
type Autocmd struct {
	ctxt *context.Context

	c  *commands.Commands
	qf []*vim.QuickfixError

	bufWritePostChan chan error
	bufWritePreChan  chan error
	wg               sync.WaitGroup

	errors []error
}

func Register(p *plugin.Plugin) {
	autocmd := new(Autocmd)
	autocmd.c = commands.NewCommands(p.Vim)

	autocmd.bufWritePreChan = make(chan error, 2)
	autocmd.bufWritePostChan = make(chan error, 2)

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmd.bufWritePre)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmd.bufWritePost)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmdVimEnter)
}
