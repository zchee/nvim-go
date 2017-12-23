// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"time"

	"nvim-go/config"
	"nvim-go/log"
	"nvim-go/nvimutil"
)

// VimEnter gets user config variables and assign to global variable when autocmd VimEnter.
func (a *Autocmd) VimEnter(cfg *config.Config) {
	defer nvimutil.Profile(time.Now(), "VimEnter")

	cfg.Global.ChannelID = a.Nvim.ChannelID()

	config.Get(a.Nvim, cfg)
	cfg2, err := config.Read()
	if err != nil {
		log.Debugln(err)
	}
	// log.DebugDump(cfg2)
	config.Merge(cfg, cfg2)

	a.buildctxt.SetContext(a.buildctxt.Dir)
}
