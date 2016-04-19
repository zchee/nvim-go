// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/garyburd/neovim-go/vim"
)

// ErrorlistData represents an item in a quickfix and locationlist.
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

// SetLoclist set the error results data to current buffer's locationlist.
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

// OpenLoclist open or close the current buffer's locationlist window.
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
		p.Command("lclose")
		if err := p.Wait(); err != nil {
			return err
		}
	}

	return nil
}

// CloseLoclist close the current buffer's locationlist window.
func CloseLoclist(v *vim.Vim) error {
	return v.Command("lclose")
}

// SetQuickfix set the error results data to quickfix list.
func SetQuickfix(p *vim.Pipeline, qflist []*ErrorlistData) error {
	p.Call("setqflist", nil, qflist)

	return nil
}

// OpenOuickfix open the quickfix list window.
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

// CloseQuickfix close the quickfix list window.
func CloseQuickfix(v *vim.Vim) error {
	return v.Command("cclose")
}

// SplitPos split the result text of the vim error list syntax.
func SplitPos(pos string) (string, int, int) {
	file := strings.Split(pos, ":")
	line, err := strconv.ParseInt(file[1], 10, 64)
	if err != nil {
		line = 0
	}
	col, err := strconv.ParseInt(file[2], 10, 64)
	if err != nil {
		col = 0
	}

	fname, err := filepath.Abs(file[0])
	if err != nil {
		return file[0], int(line), int(col)
	}

	return fname, int(line), int(col)
}

// ParseError parse a typical output of command written in Go.
func ParseError(v *vim.Vim, errors string, basedir string) []*ErrorlistData {
	var errlist []*ErrorlistData

	el := strings.Split(errors, "\n")
	for _, es := range el {
		if e := strings.SplitN(es, ":", 3); len(e) > 1 {
			line, err := strconv.ParseInt(e[1], 10, 64)
			if err != nil {
				continue
			}
			errlist = append(errlist, &ErrorlistData{
				FileName: filepath.Join(basedir, e[0]),
				LNum:     int(line),
				Text:     e[2],
			})
		}
	}
	return errlist
}
