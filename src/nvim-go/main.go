// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	_ "net/http/pprof"

	_ "nvim-go/autocmd"
	_ "nvim-go/commands"
	_ "nvim-go/commands/analyze"
	_ "nvim-go/commands/delve"

	"github.com/garyburd/neovim-go/vim/plugin"
)

func main() {
	plugin.Main()
}
