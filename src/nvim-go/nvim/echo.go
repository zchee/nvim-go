// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

// Echo provide the vim 'echo' command.
func Echo(v *vim.Vim, format string, a ...interface{}) error {
	v.Command("redraw!")
	return v.Command("echo '" + fmt.Sprintf(format, a...) + "'")
}

// Echomsg provide the vim 'echomsg' command.
func Echomsg(v *vim.Vim, a ...interface{}) error {
	v.Command("redraw!")
	return v.Command("echomsg '" + fmt.Sprintln(a...) + "'")
}

// Echoerr provide the vim 'echoerr' command.
func Echoerr(v *vim.Vim, a ...interface{}) error {
	v.Command("redraw!")
	return v.Command("echoerr '" + fmt.Sprintln(a...) + "'")
}

// Echohl provide the vim 'echohl' command with message prefix and message color highlighting.
// Nomally, used to output the any command results.
func Echohl(v *vim.Vim, prefix string, highlight string, format string, a ...interface{}) error {
	v.Command("redraw!")
	return v.Command("echo '" + prefix + "' | echohl " + highlight + " | echon '" + fmt.Sprintf(format, a...) + "' | echohl None")
}

// ReportError output of the accumulated errors report.
func ReportError(v *vim.Vim, format string, a ...interface{}) error {
	v.Command("redraw!")
	return v.ReportError(fmt.Sprintf(format, a...))
}
