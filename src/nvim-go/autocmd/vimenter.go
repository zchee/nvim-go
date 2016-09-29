// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"log"
	"net/http"
	"runtime"

	"nvim-go/config"

	vim "github.com/neovim/go-client/nvim"
)

// autocmdVimEnter wrapper vimEnter function use goroutine.
func autocmdVimEnter(v *vim.Nvim, cfg *config.Config) {
	go vimEnter(v, cfg)
}

// vimEnter gets user config variables and assign to global variable when autocmd VimEnter.
// If config.DebugPprof is true, start net/pprof debugging.
func vimEnter(v *vim.Nvim, cfg *config.Config) (err error) {
	cfg.Global.ChannelID = v.ChannelID()
	if err != nil {
		return err
	}

	config.Get(v, cfg)

	if config.DebugPprof {
		addr := "127.0.0.1:6060"
		log.Printf("Start pprof debug listen %s\n", addr)

		runtime.SetBlockProfileRate(1)
		go func() {
			log.Println(http.ListenAndServe(addr, nil))
		}()
	}

	return nil
}
