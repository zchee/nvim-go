// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"bytes"
	"fmt"
	"go/build"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nvim-go/config"
	"nvim-go/context"

	"github.com/neovim/go-client/nvim"
)

// SetLoclist set the error results data to current buffer's locationlist.
func SetLoclist(v *nvim.Nvim, loclist []*nvim.QuickfixError) error {
	// setloclist({nr}, {list} [, {action}])
	// v.Call(fname string, result interface{}, args ...interface{})
	if len(loclist) > 0 {
		v.Call("setloclist", nil, 0, loclist)
	} else {
		v.Command("lexpr ''")
	}

	return nil
}

var (
	listtype     ErrorListType
	openlistCmd  func() error
	closelistCmd func() error
	clearlistCmd func() error
	setlistCmd   func(errlist []*nvim.QuickfixError) error
)

type ErrorListType string

const (
	Quickfix     ErrorListType = "quickfix"
	LocationList               = "locationlist"
)

func getListCmd(v *nvim.Nvim) {
	if listtype == "" {
		listtype = ErrorListType(config.ErrorListType)
	}
	switch listtype {
	case Quickfix:
		openlistCmd = func() error { return v.Command("copen") }
		closelistCmd = func() error { return v.Command("cclose") }
		clearlistCmd = func() error { return v.Command("cgetexpr ''") }
		setlistCmd = func(errlist []*nvim.QuickfixError) error { return v.Call("setqflist", nil, errlist, "r") }
	case LocationList:
		openlistCmd = func() error { return v.Command("lopen") }
		closelistCmd = func() error { return v.Command("lclose") }
		clearlistCmd = func() error { return v.Command("lgetexpr ''") }
		setlistCmd = func(errlist []*nvim.QuickfixError) error { return v.Call("setloclist", nil, 0, errlist, "r") }
	}
}

// ErrorList merges the errlist map items and open the locationlist window.
// TODO(zchee): This function will reports the errors with open the quickfix window, but will close
// the quickfix window if no errors.
// Do ErrorList function name is appropriate?
func ErrorList(v *nvim.Nvim, errors map[string][]*nvim.QuickfixError, keep bool) error {
	if listtype == "" {
		getListCmd(v)
	}

	if errors == nil || len(errors) == 0 {
		defer clearlistCmd()
		return closelistCmd()
	}

	var errlist []*nvim.QuickfixError
	for _, err := range errors {
		errlist = append(errlist, err...)
	}
	if err := SetErrorlist(v, errlist); err != nil {
		return err
	}

	if keep {
		w, err := v.CurrentWindow()
		if err != nil {
			return err
		}
		defer v.SetCurrentWindow(w)
	}
	return openlistCmd()
}

// SetErrorlist set the error results data to Neovim error list.
func SetErrorlist(v *nvim.Nvim, errlist []*nvim.QuickfixError) error {
	if setlistCmd == nil {
		getListCmd(v)
	}

	return setlistCmd(errlist)
}

// ClearErrorlist clear the Neovim error list.
func ClearErrorlist(v *nvim.Nvim, close bool) error {
	if clearlistCmd == nil {
		getListCmd(v)
	}

	if close {
		defer closelistCmd()
	}
	return clearlistCmd()
}

// OpenLoclist open or close the current buffer's locationlist window.
func OpenLoclist(v *nvim.Nvim, w nvim.Window, loclist []*nvim.QuickfixError, keep bool) error {
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
func CloseLoclist(v *nvim.Nvim) error {
	return v.Command("lclose")
}

// SetQuickfix set the error results data to quickfix list.
func SetQuickfix(p *nvim.Pipeline, qflist []*nvim.QuickfixError) error {
	p.Call("setqflist", nil, qflist)

	return nil
}

// OpenOuickfix open the quickfix list window.
func OpenOuickfix(p *nvim.Pipeline, w nvim.Window, keep bool) error {
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
func CloseQuickfix(v *nvim.Nvim) error {
	return v.Command("cclose")
}

func GotoPos(v *nvim.Nvim, w nvim.Window, pos token.Position, cwd string) error {
	fname, line, col := SplitPos(pos.String(), cwd)

	v.Command(fmt.Sprintf("edit %s", fname))
	v.SetWindowCursor(w, [2]int{line, col - 1})
	defer v.Command(`lclose`)

	return v.Command(`normal zz`)
}

// SplitPos parses a string of form 'token.Pos', and return the relative
// filepath from the current working directory path.
func SplitPos(pos string, cwd string) (string, int, int) {
	position := strings.Split(pos, ":")

	fname := position[0]
	line, err := strconv.ParseInt(position[1], 10, 64)
	if err != nil {
		line = 0
	}
	col, err := strconv.ParseInt(position[2], 10, 64)
	if err != nil {
		col = 0
	}

	if strings.HasPrefix(fname, cwd) {
		frel := strings.TrimPrefix(fname, cwd+string(filepath.Separator))
		if fname != frel {
			return frel, int(line), int(col)
		}
	}

	return fname, int(line), int(col)
}

var errRe = regexp.MustCompile(`(?m)^([^:]+):(\d+)(?::(\d+))?:\s(.*)`)

// ParseError parses a typical error message of Go compile tools.
// TODO(zchee): More elegant way
func ParseError(errors []byte, cwd string, ctxt *context.Build) ([]*nvim.QuickfixError, error) {
	var (
		parentDir string
		fpath     string
		errlist   []*nvim.QuickfixError
		hasPrefix bool
	)

	// m[1]: relative file path of error file
	// m[2]: line number of error point
	// m[3]: column number of error point
	// m[4]: error description text
	for _, m := range errRe.FindAllSubmatch(errors, -1) {
		fpath = filepath.Base(string(bytes.Replace(m[1], []byte{'\t'}, nil, -1)))
		if bytes.Contains(m[1], []byte{'#'}) {
			// m[1][2:] is trim '# ' from errors directory
			// '# nvim-go/nvim' -> 'nvim-go/nvim'
			path := bytes.Split(m[1][2:], []byte{'\n'})

			// save the parent directory path for the second and subsequent error description
			parentDir = string(path[0])
			fpath = filepath.Base(string(bytes.Replace(path[1], []byte{'\t'}, nil, -1)))
			hasPrefix = true
		}

		filename := filepath.Join(parentDir, fpath)

		switch ctxt.Tool {
		case "go":
			goSrcPath := filepath.Join(build.Default.GOPATH, "src")
			currentDir := strings.TrimPrefix(cwd, goSrcPath+string(filepath.Separator))
			filename = strings.TrimPrefix(filepath.Clean(filename), currentDir+string(filepath.Separator))

		case "gb":
			if !filepath.IsAbs(filename) {
				filename = filepath.Join(ctxt.ProjectRoot, "src", filename)
			}
			if frel, err := filepath.Rel(cwd, filename); err == nil {
				filename = frel
			}
		}

		if !hasPrefix {
			filename = string(bytes.Replace(m[1], []byte{'\t'}, nil, -1))
		}

		line, err := strconv.Atoi(string(m[2]))
		if err != nil {
			return nil, err
		}

		col, err := strconv.Atoi(string(m[3]))
		// fallback if cannot convert col to type int
		// Basically, col == ""
		if err != nil {
			col = 0
		}

		errlist = append(errlist, &nvim.QuickfixError{
			FileName: filename,
			LNum:     line,
			Col:      col,
			Text:     string(bytes.TrimSpace(m[4])),
		})
	}

	return errlist, nil
}
