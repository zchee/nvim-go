package delve

import (
	"net"
	"nvim-go/nvim"
	"os/exec"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

// startServer starts the delve headless server and replace server Stdout & Stderr.
func (d *delveClient) startServer(cmd, path string) error {
	dlvBin, err := exec.LookPath("dlv")
	if err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	// TODO(zchee): costomizable build flag
	args := []string{cmd, path, "--headless=true", "--accept-multiclient=true", "--api-version=2", "--log", "--listen=" + addr}
	d.server = exec.Command(dlvBin, args...)

	d.server.Stdout = &d.serverOut
	d.server.Stderr = &d.serverErr

	if err := d.server.Start(); err != nil {
		err = errors.New(d.serverOut.String())
		d.serverOut.Reset()
		return errors.Annotate(err, "delve/server.startServer")
	}

	return nil
}

// waitServer Waits for dlv launch the headless server.
// `net.Dial` is better way?
// http://stackoverflow.com/a/30838807/5228839
func (d *delveClient) waitServer(v *vim.Vim) error {
	defer nvim.ClearMsg(v)
	nvim.EchoProgress(v, "Delve", "Wait for running dlv server")

	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		break
	}

	if err := d.setupDelveClient(v); err != nil {
		return errors.Annotate(err, "delve/server.waitServer")
	}

	return d.printLogs(v, "", []byte("Type 'help' for list of commands."))
}
