// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import "github.com/neovim-go/vim/plugin"

func Register(p *plugin.Plugin) {
	p.HandleCommand(&plugin.CommandOptions{Name: "Gobuild", Bang: true, Eval: "[getcwd(), expand('%:p:h')]"}, cmdBuild)
	p.HandleCommand(&plugin.CommandOptions{Name: "Godef", Eval: "expand('%:p:h')"}, cmdDef)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gofmt", Eval: "expand('%:p:h')"}, cmdFmt)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoGenerateTest", NArgs: "*", Complete: "file", Eval: "expand('%:p:h')"}, cmdGenerateTest)
	p.HandleFunction(&plugin.FunctionOptions{Name: "GoGuru", Eval: "[getcwd(), expand('%:p'), &modified, line2byte(line('.')) + (col('.')-2)]"}, funcGuru)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoIferr", Eval: "expand('%:p')"}, cmdIferr)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gometalinter", Eval: "getcwd()"}, cmdMetalinter)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gorename", NArgs: "?", Bang: true, Eval: "[getcwd(), expand('%:p'), expand('<cword>')]"}, cmdRename)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gorun", NArgs: "*", Eval: "expand('%:p')"}, cmdRun)
	p.HandleCommand(&plugin.CommandOptions{Name: "Gotest", NArgs: "*", Eval: "expand('%:p:h')"}, cmdTest)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoTestSwitch", Eval: "[getcwd(), expand('%:p')]"}, cmdTestSwitch)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoByteOffset", Range: "%", Eval: "expand('%:p')"}, cmdByteOffset)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoBuffers"}, cmdBuffers)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoWindows"}, cmdWindows)
	p.HandleCommand(&plugin.CommandOptions{Name: "GoTabpages"}, cmdTabpagas)
}
