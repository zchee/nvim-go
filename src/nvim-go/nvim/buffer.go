// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"encoding/binary"

	"github.com/garyburd/neovim-go/vim"
)

var (
	b vim.Buffer
	w vim.Window
)

func ByteOffset(v *vim.Vim) (int, error) {
	p := v.NewPipeline()

	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return 0, err
	}

	cursor, err := v.WindowCursor(w)
	if err != nil {
		return 0, err
	}
	byteBuf, err := v.BufferLineSlice(b, 0, -1, true, true)
	if err != nil {
		return 0, err
	}

	if cursor[0] == 1 {
		return (1 + (cursor[1] - 1)), nil
	}

	offset := 0
	line := 1
	for _, buf := range byteBuf {
		if line == cursor[0] {
			offset++
			break
		}
		offset += (binary.Size(buf) + 1)
		line++
	}

	return (offset + (cursor[1] - 1)), nil
}
