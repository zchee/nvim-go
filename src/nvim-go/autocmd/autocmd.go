// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"sync"

	"nvim-go/buildctx"
	"nvim-go/command"
	"nvim-go/logger"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"go.uber.org/zap"
	"golang.org/x/sync/syncmap"
)

// Autocmd represents a autocmd context.
type Autocmd struct {
	ctx context.Context
	log *zap.Logger

	Nvim         *nvim.Nvim
	buildContext *buildctx.Context
	cmd          *command.Command

	bufWritePostChan chan error
	bufWritePreChan  chan interface{}
	mu               sync.Mutex
	wg               sync.WaitGroup

	errs *syncmap.Map
}

// Register register autocmd to nvim.
func Register(ctx context.Context, p *plugin.Plugin, buildContext *buildctx.Context, cmd *command.Command) {
	autocmd := &Autocmd{
		ctx:              ctx,
		log:              logger.FromContext(ctx).Named("autocmd"),
		Nvim:             p.Nvim,
		buildContext:     buildContext,
		cmd:              cmd,
		bufWritePreChan:  make(chan interface{}),
		bufWritePostChan: make(chan error),
		errs:             new(syncmap.Map),
	}

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.BufEnter)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.bufWritePost)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.bufWritePre)
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.VimEnter)
}
