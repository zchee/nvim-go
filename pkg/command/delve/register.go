// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"context"

	"github.com/neovim/go-client/nvim/plugin"

	"github.com/zchee/nvim-go/pkg/buildctxt"
)

// Register register nvim-go's delve command or function to Neovim over the msgpack-rpc plugin interface.
func Register(ctx context.Context, p *plugin.Plugin, buildContext *buildctxt.Context) {
	d := NewDelve(ctx, p.Nvim, buildContext)

	// Debug compile and begin debugging program.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvDebug", NArgs: "*", Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdDebug)
	// Connect connect to a headless debug server.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvConnect", NArgs: "*", Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdConnect)

	// Breakpoint sets a breakpoint.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvBreakpoint", NArgs: "*", Eval: "[expand('%:p')]", Complete: "customlist,FunctionsCompletion"}, d.cmdBreakpoint)

	// Stepping execution control
	// Continue run until breakpoint or program termination.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvContinue", NArgs: "*", Eval: "[expand('%:p:h')]"}, d.cmdContinue)
	// Next step over to next source line.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvNext", Eval: "[expand('%:p:h')]"}, d.cmdNext)

	// restart restart the process.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvRestart"}, d.cmdRestart) // Restart process.

	// stdin interactive mode
	// TODO(zchee): Support contextual command completion
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvStdin"}, d.cmdStdin)
	// RPC export
	p.Handle("DlvStdin", d.stdin)
	// FunctionsCompletion list of functions for command completion.
	p.HandleFunction(&plugin.FunctionOptions{Name: "FunctionsCompletion"}, d.FunctionsCompletion)

	// detach exit the debugger.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvDetach"}, d.cmdDetach)

	// State (WIP: for debug)
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvState"}, d.cmdState)

	// autocmd VimLeavePre
	// FIXME(zchee): Why "[delve]*" pattern dose not handle autocmd?
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimLeavePre", Group: "nvim-go", Pattern: "*.go,terminal,context,thread"}, d.cmdDetach)
}
