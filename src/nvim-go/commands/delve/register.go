// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import "github.com/neovim-go/vim/plugin"

func Register(p *plugin.Plugin) {
	d := NewDelve(p.Vim)

	// Launch
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvDebug", Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdDebug) // Compile and begin debugging program.

	// Breakpoint
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvBreakpoint", NArgs: "*", Eval: "[expand('%:p')]", Complete: "customlist,DlvListFunctions"}, d.cmdCreateBreakpoint) // Sets a breakpoint.

	// Stepping execution control
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvContinue", Eval: "[expand('%:p:h')]"}, d.cmdContinue) // Run until breakpoint or program termination.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvNext", Eval: "[expand('%:p:h')]"}, d.cmdNext)         // Step over to next source line.
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvRestart"}, d.cmdRestart)                              // Restart process.

	// Interactive mode
	// TODO(zchee): Support contextual command completion
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvStdin"}, d.cmdStdin)
	p.HandleFunction(&plugin.FunctionOptions{Name: "DlvListFunctions"}, d.ListFunctions) // list of functions for command completion.

	// Detach
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvDetach"}, d.cmdDetach) // Exit the debugger.

	// RPC Exports
	p.Handle("DlvStdin", d.stdin)

	// State (WIP: for debug)
	p.HandleCommand(&plugin.CommandOptions{Name: "DlvState"}, d.cmdState)

	// autocmd VimLeavePre
	// FIXME(zchee): Why "[delve]*" pattern dose not handle autocmd?
	p.HandleAutocmd(&plugin.AutocmdOptions{Event: "VimLeavePre", Group: "nvim-go", Pattern: "*.go,terminal,context,thread"}, d.cmdDetach)
}
