// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import vim "github.com/neovim/go-client/nvim"

func (d *Delve) readServerStdout(v *vim.Nvim, cmd, args string) error {
	command := cmd + " " + args

	return d.printTerminal(command, d.serverOut.Bytes())
}

func (d *Delve) readServerStderr(v *vim.Nvim, cmd, args string) error {
	command := cmd + " " + args

	return d.printTerminal(command, d.serverErr.Bytes())
}
