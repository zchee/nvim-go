// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command nvimgo is a Neovim remote plogin.
package main

import (
	"log"
	_ "net/http/pprof"
	"os"

	"github.com/garyburd/neovim-go/vim/plugin"

	_ "nvim-go/autocmd"
	_ "nvim-go/commands"
	_ "nvim-go/nvim"
	_ "nvim-go/pkgs"
	_ "nvim-go/vars"
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
