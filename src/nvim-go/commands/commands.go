// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"nvim-go/context"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

type Commands struct {
	Nvim     *nvim.Nvim
	Pipeline *nvim.Pipeline
	Batch    *nvim.Batch

	ctx *context.Context
}

func NewCommands(v *nvim.Nvim, ctx *context.Context) *Commands {
	return &Commands{
		Nvim: v,
		ctx:  ctx,
	}
}

func Register(p *plugin.Plugin, ctxt *context.Context) *Commands {
	c := NewCommands(p.Nvim, ctxt)

	// Register command and function
	// CommandOptions order: Name, NArgs, Range, Count, Addr, Bang, Register, Eval, Bar, Complete
	p.HandleCommand(&plugin.CommandOptions{Name: "Gobuild", Bang: true, Eval: "[getcwd(), expand('%:p:h')]"}, c.cmdBuild)
	p.HandleCommand(&plugin.CommandOptions{Name: "Godef", Eval: "expand('%:p:h')"}, c.cmdDef)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gofmt", Eval: "expand('%:p:h')"}, c.cmdFmt)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoGenerateTest", NArgs: "*", Range: "%", Addr: "line", Bang: true, Eval: "expand('%:p:h')", Complete: "file"}, c.cmdGenerateTest)
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoGuru", Eval: "[getcwd(), expand('%:p'), &modified, line2byte(line('.')) + (col('.')-2)]"}, c.funcGuru)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoIferr", Eval: "expand('%:p')"}, c.cmdIferr)
	p.HandleCommand(&plugin.CommandOptions{Name: "Golint", NArgs: "?", Eval: "expand('%:p')", Complete: "customlist,GoLintCompletion"}, c.cmdLint)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gometalinter", Eval: "getcwd()"}, c.cmdMetalinter)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gorename", NArgs: "?", Bang: true, Eval: "[getcwd(), expand('%:p'), expand('<cword>')]"}, c.cmdRename)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gorun", NArgs: "*", Eval: "expand('%:p')"}, c.cmdRun)
	p.HandleCommand(&plugin.CommandOptions{Name: "GorunLast", Eval: "expand('%:p')"}, c.cmdRunLast)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gotest", NArgs: "*", Eval: "expand('%:p:h')"}, c.cmdTest)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoSwitchTest", Eval: "[getcwd(), expand('%:p'), line2byte(line('.')) + (col('.')-2)]"}, c.cmdSwitchTest)
	p.HandleCommand(&plugin.CommandOptions{Name: "Govet", NArgs: "*", Eval: "[getcwd(), expand('%:p')]", Complete: "customlist,GoVetCompletion"}, c.cmdVet)

	// Commnad completion
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoLintCompletion", Eval: "getcwd()"}, c.cmdLintComplete) // list the file, directory and go packages
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoVetCompletion", Eval: "getcwd()"}, c.cmdVetComplete)   // flag for go tool vet

	// for debug
	p.HandleCommand(&plugin.CommandOptions{Name: "GoByteOffset", Range: "%", Eval: "expand('%:p')"}, c.cmdByteOffset)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoBuffers"}, c.cmdBuffers)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoWindows"}, c.cmdWindows)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoTabpages"}, c.cmdTabpagas)

	return c
}
