// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quickfix

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nvim-go/context"

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
func SetLoclist(v *vim.Vim, loclist []*ErrorlistData) error {
	// setloclist({nr}, {list} [, {action}])
	// v.Call(fname string, result interface{}, args ...interface{})
	if len(loclist) > 0 {
		v.Call("setloclist", nil, 0, loclist)
	} else {
		v.Command("lexpr ''")
	}

	return nil
}

// OpenLoclist open or close the current buffer's locationlist window.
func OpenLoclist(v *vim.Vim, w vim.Window, loclist []*ErrorlistData, keep bool) error {
	if len(loclist) == 0 {
		return v.Command("lclose")
	}

	v.Command("lopen")
	if keep {
		return v.SetCurrentWindow(w)
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

// ParseError parses a typical error message of Go compile tools.
// Error sample:
//  # nvim-go/nvim
//  echo.go:79: syntax error: non-declaration statement outside function body
//  # nvim-go/nvim/quickfix
//  locationlist.go:152: syntax error: unexpected case, expecting }
//  locationlist.go:160: syntax error: non-declaration statement outside function body
// TODO(zchee): More better regexp pattern and single for loop if possible.
func ParseError(errors []byte, cwd string, ctxt *context.Build) ([]*ErrorlistData, error) {
	var (
		errlist      []*ErrorlistData
		reErrPattern = regexp.MustCompile(`(?m)^#\s([-_./\w]+)\n([.,-_'":\s\w]+)`)
		reFile       = regexp.MustCompile(`([.\w]+):(\d+)(?::(\d+))?:\s(.*)`)
	)

	for _, m := range reErrPattern.FindAllSubmatch(errors, -1) {
		parent := string(m[1])
		errFile := m[2]

		for _, mm := range reFile.FindAllSubmatch(errFile, -1) {
			var fname string
			fpath := filepath.Join(parent, string(mm[1]))

			switch ctxt.Tool {
			case "go":
				sep := filepath.Join(ctxt.GOPATH, "src")
				c := strings.TrimPrefix(cwd, sep)
				fname = strings.TrimPrefix(filepath.Clean(fpath), c+string(filepath.Separator))

			case "gb":
				if !filepath.IsAbs(fpath) {
					fpath = filepath.Join(ctxt.ProjectDir, "src", fpath)
				}
				fname, _ = filepath.Rel(cwd, fpath)
			}

			line, _ := strconv.Atoi(string(mm[2]))
			col, _ := strconv.Atoi(string(mm[3]))

			errlist = append(errlist, &ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     string(bytes.TrimSpace(mm[4])),
			})
		}
	}

	return errlist, nil
}
