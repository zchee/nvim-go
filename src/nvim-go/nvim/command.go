// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

func Echo(v *vim.Vim, format string, a ...interface{}) error {
	return v.Command("echo '" + fmt.Sprintf(format, a...) + "'")
}

func Echomsg(v *vim.Vim, format string, a ...interface{}) error {
	return v.Command("echomsg '" + fmt.Sprintf(format, a...) + "'")
}

func Echoerr(v *vim.Vim, format string, a ...interface{}) error {
	return v.Command("echoerr '" + fmt.Sprintf(format, a...) + "'")
}

func ReportError(v *vim.Vim, format string, a ...interface{}) error {
	return v.ReportError(fmt.Sprintf(format, a...))
}
