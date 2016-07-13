// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"log"

	"github.com/juju/errors"
	"github.com/neovim-go/vim"
)

func (d *Delve) cmdDetach(v *vim.Vim) {
	go d.detach(v)
}

func (d *Delve) detach(v *vim.Vim) error {
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

func (d *Delve) kill() error {
	if d.server != nil {
		err := d.server.Process.Kill()
		if err != nil {
			return errors.Annotate(err, pkgDelve)
		}
		log.Printf("Killed delve server\n")
	}

	return nil
}
