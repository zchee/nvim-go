// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"github.com/neovim/go-client/nvim"
)

// TabpageContext represents a Neovim tabpage context.
type TabpageContext struct {
	nvim.Tabpage
}
