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
func Register(ctx context.Context, p *plugin.Plugin, buildContext *buildctxt.Context, cmd *command.Command) {
	log := logger.FromContext(ctx).Named("autocmd")
	ctx = logger.NewContext(ctx, log)

	autocmd := &Autocmd{
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
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufEnter", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, func(eval *bufEnterEval) { autocmd.BufEnter(ctx, eval) })

	// BufNewFile: Handle create the new file.
	// BufReadPre: Handle the before the read to file.
	// If create the new file, does not run the 'BufReadPre', Instead of 'BufNewFile'.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufNewFile,BufReadPre", Group: "nvim-go", Pattern: "*", Eval: "*"}, func(eval *bufReadPreEval) { autocmd.BufReadPre(ctx, eval) })

	// p.HandleAutocmd(&plugin.AutocmdOptions{Event: "WinEnter", Group: "nvim-go", Pattern: "*.go", Eval: "*"}, func(eval *winEnterEval) { autocmd.WinEnter(ctx, eval) })

	// Handle the before the write to file.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePre", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, func(eval *bufWritePreEval) { autocmd.BufWritePre(ctx, eval) })

	// Handle the after the write to file.
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "BufWritePost", Pattern: "*.go", Group: "nvim-go", Eval: "*"}, func(eval *bufWritePostEval) { autocmd.BufWritePost(ctx, eval) })

	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimLeavePre", Pattern: "*.go", Group: "nvim-go"}, func() { autocmd.VimLeavePre(ctx) })
}
