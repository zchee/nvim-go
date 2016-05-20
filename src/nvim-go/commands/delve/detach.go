package delve

import (
	"log"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

func (d *delveClient) cmdDetach(v *vim.Vim) {
	go d.detach(v)
}

func (d *delveClient) detach(v *vim.Vim) error {
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

func (d *delveClient) kill() error {
	if d.server != nil {
		err := d.server.Process.Kill()
		if err != nil {
			return errors.Annotate(err, pkgDelve)
		}
		log.Printf("Killed delve server\n")
	}

	return nil
}
