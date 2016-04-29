// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	_ "nvim-go/autocmd"
	_ "nvim-go/commands"
	_ "nvim-go/config"
	_ "nvim-go/context"
	_ "nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	if lf := os.Getenv("NEOVIM_GO_LOG_FILE"); lf != "" {
		f, err := os.OpenFile(lf, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(f)
	}
}

func main() {
	plugin.Main()
}
