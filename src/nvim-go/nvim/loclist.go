// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"go/scanner"

	"github.com/garyburd/neovim-go/vim"
)

// QuickfixError represents an item in a quickfix list.
type LoclistError struct {
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

func LoclistErrors(v *vim.Vim, b vim.Buffer, errors error) error {
	var loclist []*LoclistError

	if e, ok := errors.(scanner.Error); ok {
		loclist = append(loclist, &LoclistError{
			LNum: e.Pos.Line,
			Col:  e.Pos.Column,
			Text: e.Msg,
		})
	} else if el, ok := errors.(scanner.ErrorList); ok {
		for _, e := range el {
			loclist = append(loclist, &LoclistError{
				LNum: e.Pos.Line,
				Col:  e.Pos.Column,
				Text: e.Msg,
			})
		}
	}

	if len(loclist) == 0 {
		return errors
	}

	bufnr, err := v.BufferNumber(b)
	if err != nil {
		return err
	}
	for i := range loclist {
		loclist[i].Bufnr = bufnr
	}

	// setloclist({nr}, {list} [, {action}])
	// Call(fname string, result interface{}, args ...interface{}) error
	if err := v.Call("setloclist", nil, 0, loclist); err != nil {
		return err
	}

	return v.Command("lopen")
}

func LoclistClose(v *vim.Vim) error {
	return v.Command("lclose")
}
