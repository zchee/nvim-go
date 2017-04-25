// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"sync"

	"nvim-go/command"
	"nvim-go/ctx"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"golang.org/x/sync/syncmap"
)

// Autocmd represents a autocmd context.
type Autocmd struct {
	Nvim *nvim.Nvim
	ctx  *ctx.Context
	cmd  *command.Command

	bufWritePostChan chan error
	bufWritePreChan  chan interface{}
	mu               sync.Mutex
	wg               sync.WaitGroup

	errs *syncmap.Map
}

// Register register autocmd to nvim.
func Register(p *plugin.Plugin, ctx *ctx.Context, cmd *command.Command) {
	autocmd := &Autocmd{
		Nvim:             p.Nvim,
		ctx:              ctx,
		cmd:              cmd,
		bufWritePreChan:  make(chan interface{}),
		bufWritePostChan: make(chan error),
		errs:             new(syncmap.Map),
	}

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.BufEnter)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmd.BufWritePost)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmd.BufWritePre)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.VimEnter)
}
