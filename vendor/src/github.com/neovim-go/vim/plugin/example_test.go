// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin_test

import (
	"strings"

	"github.com/neovim-go/vim/plugin"
)

// This plugin adds the Hello function to Neovim.
func Example() {
	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleFunction(&plugin.FunctionOptions{Name: "Hello"}, func(args []string) (string, error) {
			return "Hello, " + strings.Join(args, " "), nil
		})
		return nil
	})
}
