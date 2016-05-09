// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quickfix

import (
	"bytes"
	"nvim-go/context"
	"path/filepath"
	"regexp"
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

// SplitPos parses a string of form 'token.Pos', and return the relative
// filepath from the current working directory path.
func SplitPos(pos string, cwd string) (string, int, int) {
	slc := strings.Split(pos, ":")
	line, err := strconv.ParseInt(slc[1], 10, 64)
	if err != nil {
		line = 0
	}
	col, err := strconv.ParseInt(slc[2], 10, 64)
	if err != nil {
		col = 0
	}

	fname := slc[0]
	frel := strings.TrimPrefix(fname, cwd+string(filepath.Separator))
	if fname == frel {
		return fname, int(line), int(col)
	}

	return frel, int(line), int(col)
}

// ParseError parse a typical output of command written in Go.
func ParseError(errors []byte, cwd string, ctxt *context.Build) ([]*ErrorlistData, error) {
	var (
		errlist []*ErrorlistData
		errPat  = regexp.MustCompile(`^# ([^:]+):(\d+)(?::(\d+))?:\s(.*)`)
		fname   string
	)

	for _, m := range errPat.FindAllSubmatch(errors, -1) {
		fb := bytes.Split(m[1], []byte("\n"))
		fs := string(bytes.Join(fb, []byte(string(filepath.Separator))))
		if ctxt.Tool == "go" {
			sep := ctxt.GOPATH + string(filepath.Separator) + "src" + string(filepath.Separator)
			c := strings.TrimPrefix(cwd, sep)

			fname = strings.TrimPrefix(filepath.Clean(fs), c+string(filepath.Separator))
		} else if ctxt.Tool == "gb" {
			sep := filepath.Base(cwd) + string(filepath.Separator)
			fname = strings.TrimPrefix(filepath.Clean(fs), sep)
		}

		line, _ := strconv.Atoi(string(m[2]))
		col, _ := strconv.Atoi(string(m[3]))

		errlist = append(errlist, &ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     string(bytes.TrimSpace(m[4])),
		})
	}

	return errlist, nil
}
