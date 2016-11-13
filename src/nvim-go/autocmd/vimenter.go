// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"log"
	"net/http"
	"runtime"

	"nvim-go/config"

	"github.com/neovim/go-client/nvim"
)

// autocmdVimEnter wrapper vimEnter function use goroutine.
func (a *Autocmd) VimEnter(v *nvim.Nvim, cfg *config.Config) {
	go a.vimEnter(v, cfg)
}

// vimEnter gets user config variables and assign to global variable when autocmd VimEnter.
// If config.DebugPprof is true, start net/pprof debugging.
func (a *Autocmd) vimEnter(v *nvim.Nvim, cfg *config.Config) (err error) {
	cfg.Global.ChannelID = v.ChannelID()

	a.mu.Lock()
	go func() {
		defer a.mu.Unlock()

		config.Get(v, cfg)

		if config.DebugPprof {
			addr := "127.0.0.1:6060"
			log.Printf("Start pprof debug listen %s\n", addr)

			runtime.SetBlockProfileRate(1)
			go func() {
				log.Println(http.ListenAndServe(addr, nil))
			}()
		}
	}()

	return nil
}
