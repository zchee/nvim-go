// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"fmt"

	"github.com/neovim/go-client/nvim"
)

func (d *Delve) readServerStdout(v *nvim.Nvim, cmd, args string) error {
	return d.printTerminal(fmt.Sprintf("%s %s", cmd, args), d.serverOut.Bytes())
}

func (d *Delve) readServerStderr(v *nvim.Nvim, cmd, args string) error {
	return d.printTerminal(fmt.Sprintf("%s %s", cmd, args), d.serverErr.Bytes())
}
