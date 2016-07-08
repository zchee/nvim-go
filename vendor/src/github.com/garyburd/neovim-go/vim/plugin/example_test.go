// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin_test

import (
	"strings"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleFunction("Hello", nil, func(v *vim.Vim, args []string) (string, error) {
		return "Hello, " + strings.Join(args, " "), nil
	})
}

// This plugin adds the Hello function to Neovim.
func Example() {
	plugin.Main()
}
