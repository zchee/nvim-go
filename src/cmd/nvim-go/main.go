// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main // import "github.com/zchee/nvim-go/src/cmd/nvim-go"

import (
	// For pprof debugging.
	_ "net/http/pprof"

	"nvim-go/autocmd"
	"nvim-go/commands"
	"nvim-go/commands/delve"
	"nvim-go/context"

	"github.com/neovim/go-client/nvim/plugin"
)

func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		ctxt := context.NewContext()

		c := commands.Register(p, ctxt)
		delve.Register(p, ctxt)

		autocmd.Register(p, ctxt, c)

		return nil
	})
}
