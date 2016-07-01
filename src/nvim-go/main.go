// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	// For pprof debugging.
	_ "net/http/pprof"

	// Register autocmd
	_ "nvim-go/autocmd"
	// Register commands
	_ "nvim-go/commands"
	// Register delve command and autocmd
	_ "nvim-go/commands/delve"

	"github.com/garyburd/neovim-go/vim/plugin"
)

func main() {
	plugin.Main()
}
