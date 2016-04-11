// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

func Debugf(v *vim.Vim, enable bool, format string, a ...interface{}) {
	if enable {
		// v.Command("echo '" + fmt.Sprintf(format, a...) + "'")
		fmt.Printf(format, a...)
	}
}

func Debugln(v *vim.Vim, enable bool, a ...interface{}) {
	if enable {
		// v.Command("echo '" + fmt.Sprintln(a...) + "'")
		fmt.Println(a...)
	}
}
