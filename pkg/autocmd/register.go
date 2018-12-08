// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"sync"

	"github.com/neovim/go-client/nvim/plugin"
	"go.uber.org/zap"
	"golang.org/x/exp/errors/fmt"

	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimctx"
)

// Register registers autocmd to Neovim.
//
//  :help rpcrequest()
//  :help rpcnotify()
func Register(pctx context.Context, cancel func(), p *plugin.Plugin, buildContext *buildctxt.Context, cmd *command.Command) {
	log := logger.FromContext(pctx).Named("autocmd")
	ctx := logger.NewContext(pctx, log)

	autocmd := &Autocmd{
		ctx:              ctx,
		cancel:           cancel,
		Nvim:             p.Nvim,
		buildContext:     buildContext,
		cmd:              cmd,
		bufWritePreChan:  make(chan interface{}),
		bufWritePostChan: make(chan error),
		errs:             new(sync.Map),
	}

	// Handle the initial start Neovim process.
	// Note that does not run the 'VimEnter' handler if open the *not* go file. Because 'VimEnter' handler already run the other file or directory.
	// TODO(zchee): consider Pattern to '*' instead of '*.go' with get '&filetype' and early return
	// p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimEnter", Pattern: "*", Group: "nvim-go", Eval: "*"}, autocmd.VimEnter)

	p.Handle(nvimctx.Method, func(event ...interface{}) {
		log.Debug(fmt.Sprintf("handles %s", nvimctx.Method), zap.Any("event", event))
	})

	// Handle the open the file.
	// If open the file at first, run the 'BufEnter' -> 'VimEnter'.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.BufEnter)

	// BufNewFile: Handle create the new file.
	// BufReadPre: Handle the before the read to file.
	// If create the new file, does not run the 'BufReadPre', Instead of 'BufNewFile'.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufNewFile,BufReadPre", Group: "nvim-go-autocmd", Pattern: "*.go", Eval: "*"}, autocmd.BufReadPre)

	// p.HandleAutocmd(&plugin.AutocmdOptions{Event: "WinEnter", Group: "nvim-go-autocmd", Pattern: "*.go", Eval: "*"}, autocmd.WinEnter)

	// Handle the before the write to file.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.BufWritePre)

	// Handle the after the write to file.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmd.bufWritePost)

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimLeavePre", Pattern: "*.go", Group: "nvim-go"}, autocmd.VimLeavePre)
}
