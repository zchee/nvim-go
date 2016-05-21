// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"log"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

func (d *delve) cmdDetach(v *vim.Vim) {
	go d.detach(v)
}

func (d *delve) detach(v *vim.Vim) error {
	defer d.kill()
	if d.processPid != 0 {
		err := d.client.Detach(true)
		if err != nil {
			return errors.Annotate(err, pkgDelve)
		}
		log.Printf("Detached delve client\n")
	}

	return nil
}

func (d *delve) kill() error {
	if d.server != nil {
		err := d.server.Process.Kill()
		if err != nil {
			return errors.Annotate(err, pkgDelve)
		}
		log.Printf("Killed delve server\n")
	}

	return nil
}
