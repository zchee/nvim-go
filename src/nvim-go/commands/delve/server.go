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

type Config struct {
	addr  string
	flags []string
	path  string
	pid   int
}

// startServer starts the delve headless server and replace server Stdout & Stderr.
func (d *Delve) startServer(cmd string, cfg Config) error {
	dlv, err := exec.LookPath("dlv")
	if err != nil {
		return errors.WithStack(err)
	}

	switch cmd {
	case "attach":
		// TODO(zchee): implements
	case "connect":
		// connect command must be addr to the second argument
		d.server = exec.Command(dlv, cmd, cfg.addr, "--log")
	case "debug":
		// debug command must be package path to the second argument, and need "--accept-multiclient" flag
		d.server = exec.Command(dlv, cmd, cfg.path, "--headless", "--listen="+cfg.addr, "--accept-multiclient", "--api-version=2", "--log")
	case "exec":
		// TODO(zchee): implements
	case "test":
		// TODO(zchee): implements
	case "trace":
		// TODO(zchee): implements
	}
	// append other flags such as build flags
	d.server.Args = append(d.server.Args, cfg.flags...)

	if err := d.server.Start(); err != nil {
		err = errors.New(d.serverOut.String())
		d.serverOut.Reset()
		return errors.WithStack(err)
	}

	return nil
}

// dialServer dial the dlv launch the headless server.
// `net.Dial` is better way?
// http://stackoverflow.com/a/30838807/5228839
func (d *Delve) dialServer(v *nvim.Nvim, addr string) error {
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

	return nil
}
