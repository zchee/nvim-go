// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import "github.com/neovim-go/vim"

func (d *Delve) readServerStdout(v *vim.Vim, cmd, args string) error {
	command := cmd + " " + args

	return d.printTerminal(command, d.serverOut.Bytes())
}

func (d *Delve) readServerStderr(v *vim.Vim, cmd, args string) error {
	command := cmd + " " + args

	return d.printTerminal(command, d.serverErr.Bytes())
}
