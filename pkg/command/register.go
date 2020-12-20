// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/logger"
)

// Register register nvim-go command or function to Neovim over the msgpack-rpc plugin interface.
func Register(ctx context.Context, p *plugin.Plugin, bctxt *buildctxt.Context) *Command {
	c := NewCommand(ctx, p.Nvim, bctxt)
	log := logger.FromContext(ctx).Named("command")
	ctx = logger.NewContext(ctx, log)

	// CommandOptions order:
	//  Name, NArgs, Range, Count, Addr, Bang, Register, Eval, Bar, Complete
	p.HandleCommand(&plugin.CommandOptions{Name: "GoBuild", NArgs: "*", Bang: true, Eval: "[getcwd(), expand('%:p')]"},
		func(args []string, bang bool, eval *CmdBuildEval) {
			c.cmdBuild(ctx, args, bang, eval)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoCover", NArgs: "?", Eval: "[getcwd(), expand('%:p')]"},
		func(args []string, eval *cmdCoverEval) {
			c.cmdCover(ctx, args, eval)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoCoverClear"},
		func() {
			c.cmdClearCover(ctx)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoFmt", Eval: "expand('%:p:h')"},
		func(dir string) {
			c.cmdFmt(ctx, dir)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoGenerateTest", NArgs: "*", Range: "%", Addr: "line", Bang: true, Eval: "expand('%:p:h')", Complete: "file"},
		func(args []string, ranges [2]int, bang bool, dir string) {
			c.cmdGenerateTest(ctx, args, ranges, bang, dir)
		})
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoGuru", Eval: "[getcwd(), expand('%:p'), &modified, line2byte(line('.')) + (col('.')-2)]"},
		func(args []string, eval *funcGuruEval) {
			c.funcGuru(ctx, args, eval)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoIferr", Eval: "expand('%:p')"},
		func(file string) {
			c.cmdIferr(ctx, file)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoLint", NArgs: "?", Eval: "expand('%:p')", Complete: "customlist,GoLintCompletion"},
		func(args []string, file string) {
			c.cmdLint(ctx, args, file)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoMetalinter", Eval: "getcwd()"},
		func(cwd string) {
			c.cmdMetalinter(ctx, cwd)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoRename", NArgs: "?", Bang: true, Eval: "[getcwd(), expand('%:p'), expand('<cword>')]"},
		func(args []string, bang bool, eval *cmdRenameEval) {
			c.cmdRename(ctx, args, bang, eval)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoRun", NArgs: "*", Eval: "expand('%:p')"},
		func(args []string, file string) {
			c.cmdRun(ctx, args, file)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoRunLast", Eval: "expand('%:p')"},
		func(file string) {
			c.cmdRunLast(ctx, file)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoTest", NArgs: "*", Eval: "expand('%:p:h')"},
		func(args []string, dir string) {
			c.cmdTest(ctx, args, dir)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoSwitchTest", Eval: "*"},
		func(eval *cmdTestSwitchEval) {
			c.SwitchTest(ctx, eval)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoVet", NArgs: "*", Eval: "[getcwd(), expand('%:p')]", Complete: "customlist,GoVetCompletion"},
		func(args []string, eval *CmdVetEval) {
			c.cmdVet(ctx, args, eval)
		})

	// Commnad completion
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoLintCompletion", Eval: "getcwd()"}, // list the file, directory and go packages
		func(a *nvim.CommandCompletionArgs, cwd string) {
			c.cmdLintComplete(ctx, a, cwd)
		})
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoVetCompletion", Eval: "getcwd()"}, // flag for go tool vet
		func(a *nvim.CommandCompletionArgs, dir string) {
			c.cmdVetComplete(ctx, a, dir)
		})

	// for debug
	p.HandleCommand(&plugin.CommandOptions{Name: "GoByteOffset", Eval: "expand('%:p')"},
		func() {
			c.cmdByteOffset(ctx)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoBuffers"},
		func() {
			c.cmdBuffers(ctx)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoWindows"},
		func() {
			c.cmdWindows(ctx)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoTabpages"},
		func() {
			c.cmdTabpagas(ctx)
		})
	p.HandleCommand(&plugin.CommandOptions{Name: "GoNotify", NArgs: "*"},
		func(args []string) {
			c.cmdNotify(ctx, args)
		})

	return c
}
