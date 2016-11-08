// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"net"
	"os/exec"

	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

// startServer starts the delve headless server and replace server Stdout & Stderr.
func (d *Delve) startServer(cmd, path string) error {
	dlvBin, err := exec.LookPath("dlv")
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	// TODO(zchee): costomizable build flag
	args := []string{cmd, path, "--headless=true", "--accept-multiclient=true", "--api-version=2", "--log", "--listen=" + addr}
	d.server = exec.Command(dlvBin, args...)

	d.server.Stdout = &d.serverOut
	d.server.Stderr = &d.serverErr

	if err := d.server.Start(); err != nil {
		err = errors.New(d.serverOut.String())
		d.serverOut.Reset()
		return errors.Wrap(err, "delve/server.startServer")
	}

	return nil
}

// waitServer Waits for dlv launch the headless server.
// `net.Dial` is better way?
// http://stackoverflow.com/a/30838807/5228839
func (d *Delve) waitServer(v *nvim.Nvim) error {
	defer nvimutil.EchohlAfter(v, "Delve", nvimutil.ProgressColor, "Ready")
	nvimutil.EchoProgress(v, "Delve", "Wait for running dlv server")

	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		break
	}

	if err := d.setupDelve(v); err != nil {
		return errors.Wrap(err, "delve/server.waitServer")
	}

	return d.printTerminal("", []byte("Type 'help' for list of commands."))
}
