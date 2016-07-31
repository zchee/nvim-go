// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"sync"

	"nvim-go/commands"
	"nvim-go/context"

	"github.com/neovim-go/vim/plugin"
)

// Autocmd represents a autocmd context.
type Autocmd struct {
	ctxt *context.Context
	c    *commands.Commands

	bufWritePostChan chan error
	bufWritePreChan  chan interface{}
	wg               sync.WaitGroup

	errors []error
}

func Register(p *plugin.Plugin, ctxt *context.Context, c *commands.Commands) {
	autocmd := new(Autocmd)
	autocmd.ctxt = ctxt
	autocmd.c = c

	autocmd.bufWritePreChan = make(chan interface{})
	autocmd.bufWritePostChan = make(chan error)

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmd.cmdBufWritePre)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmd.cmdBufWritePost)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmdVimEnter)
}
