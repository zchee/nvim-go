// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"log"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

func (d *Delve) cmdDetach(v *nvim.Nvim) {
	go d.detach(v)
}

func (d *Delve) detach(v *nvim.Nvim) error {
	defer d.kill()
	if d.processPid != 0 {
		err := d.client.Detach(true)
		if err != nil {
			return nvimutil.ErrorWrap(d.Nvim, errors.WithStack(err))
		}
		log.Printf("Detached delve client\n")
	}

	return nil
}

func (d *Delve) kill() error {
	if d.server != nil {
		err := d.server.Process.Kill()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Printf("Killed delve server\n")
	}

	return nil
}
