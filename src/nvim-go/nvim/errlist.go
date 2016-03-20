// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"strconv"
	"strings"

	"github.com/garyburd/neovim-go/vim"
)

// Loclist represents an item in a quickfix list.
type ErrorlistData struct {
	// Buffer number
	Bufnr int `msgpack:"bufnr,omitempty"`

	// Name of a file; only used when bufnr is not present or it is invalid.
	FileName string `msgpack:"filename,omitempty"`

	// Line number in the file.
	LNum int `msgpack:"lnum,omitempty"`

	// Column number (first column is 1).
	Col int `msgpack:"col,omitempty"`

	// When Vcol is != 0,  Col is visual column.
	VCol int `msgpack:"vcol,omitempty"`

	// Error number.
	Nr int `msgpack:"nr,omitempty"`

	// Search pattern used to locate the error.
	Pattern string `msgpack:"pattern,omitempty"`

	// Description of the error.
	Text string `msgpack:"text,omitempty"`

	// Single-character error type, 'E', 'W', etc.
	Type string `msgpack:"type,omitempty"`

	// Valid is non-zero if this is a recognized error message.
	Valid int `msgpack:"valid,omitempty"`
}

func SetLoclist(p *vim.Pipeline, loclist []*ErrorlistData) error {
	// setloclist({nr}, {list} [, {action}])
	// Call(fname string, result interface{}, args ...interface{})
	if len(loclist) > 0 {
		p.Call("setloclist", nil, 0, loclist)
	} else {
		p.Command("lexpr ''")
	}

	return nil
}

func OpenLoclist(p *vim.Pipeline, w vim.Window, loclist []*ErrorlistData, keep bool) error {
	if len(loclist) > 0 {
		p.Command("lopen")
		if keep {
			p.SetCurrentWindow(w)
		}
		if err := p.Wait(); err != nil {
			return err
		}
	} else {
		p.Command("redraw!")
		p.Command("lclose")
		if err := p.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func CloseLoclist(v *vim.Vim) error {
	return v.Command("lclose")
}

func SetQuickfix(p *vim.Pipeline, qflist []*ErrorlistData) error {
	p.Call("setqflist", nil, qflist)

	return nil
}

func OpenOuickfix(p *vim.Pipeline, w vim.Window, keep bool) error {
	p.Command("copen")
	if keep {
		p.SetCurrentWindow(w)
	}
	if err := p.Wait(); err != nil {
		return err
	}

	return nil
}

func CloseQuickfix(v *vim.Vim) error {
	return v.Command("cclose")
}

func SplitPos(pos string) (string, int, int) {
	file := strings.Split(pos, ":")
	line, _ := strconv.ParseInt(file[1], 10, 64)
	col, _ := strconv.ParseInt(file[2], 10, 64)

	return file[0], int(line), int(col)
}
