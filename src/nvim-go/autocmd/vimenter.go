// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"log"
	"net/http"

	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("VimEnter",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "*"}, autocmdVimEnter)
}

func autocmdVimEnter(v *vim.Vim, cfg *config.Config) {
	go vimEnter(v, cfg)
}

func vimEnter(v *vim.Vim, cfg *config.Config) error {
	cfg.Remote.ChannelID, _ = v.ChannelID()
	config.Getconfig(v, cfg)

	if config.DebugPprof {
		log.Printf("Start pprof debug\n")
		log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
	}
	return nil
}
