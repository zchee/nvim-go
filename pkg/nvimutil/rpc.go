// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"github.com/neovim/go-client/nvim"

	"github.com/zchee/nvim-go/pkg/config"
)

func Notify(n *nvim.Nvim, method string, args ...string) error {
	return n.Call("rpcnotify", nil, config.ChannelID, method, args)
}
