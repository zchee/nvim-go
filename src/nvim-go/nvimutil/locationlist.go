// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"bytes"

	"fmt"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nvim-go/config"
	"nvim-go/ctx"
	"nvim-go/internal/pathutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
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

// ErrorListType represents a neovim error list type.
type ErrorListType string

const (
	// Quickfix quickfix error list type.
	Quickfix ErrorListType = "quickfix"
	// LocationList locationlist error list type.
	LocationList = "locationlist"
)

func getListCmd(v *nvim.Nvim) {
	listtype = ErrorListType(config.ErrorListType)
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
func SetQuickfix(b *nvim.Batch, qflist []*nvim.QuickfixError) error {
	b.Call("setqflist", nil, qflist)

	return nil
}

// OpenOuickfix open the quickfix list window.
func OpenOuickfix(b *nvim.Batch, w nvim.Window, keep bool) error {
	b.Command("copen")
	if keep {
		b.SetCurrentWindow(w)
	}
	if err := b.Execute(); err != nil {
		return err
	}

	return nil
}

// CloseQuickfix close the quickfix list window.
func CloseQuickfix(v *nvim.Nvim) error {
	return v.Command("cclose")
}

// GotoPos change current buffer from the pos position with 'zz' normal behavior.
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

// regexp pattern: https://regex101.com/r/bUVZpH/2
var errRe = regexp.MustCompile(`(?m)^(?:#\s([[:graph:]]+))?(?:[\s\t]+)?([^\s:]+):(\d+)(?::(\d+))?(?::)?\s(.*)`)

// ParseError parses a typical Go tools error messages.
func ParseError(errs []byte, cwd string, buildContext *ctx.Build, ignoreDirs []string) ([]*nvim.QuickfixError, error) {
	var (
		// packagePath for the save the error files parent directory.
		// It will be re-assigned if "# " is in the error message.
		packagePath string
		errlist     []*nvim.QuickfixError
	)

	// m[1]: package path with "# " prefix
	// m[2]: error files relative path
	// m[3]: line number of error point
	// m[4]: column number of error point
	// m[5]: error description text
	for _, m := range errRe.FindAllSubmatch(errs, -1) {
		if m[1] != nil {
			// Save the package path for the second subsequent errors
			packagePath = string(m[1])
		}
		filename := string(m[2])

		// Avoid the local package error. like "package foo" and edit "cmd/foo/main.go"
		if !strings.Contains(filename, "../") && (!filepath.IsAbs(filename) && packagePath != "") {
			// Joins the packagePath and error file
			filename = filepath.Join(packagePath, filepath.Base(filename))
		}

		// Cleanup filename to relative path of current working directory
		switch buildContext.Tool {
		case "go":
			var sep string
			switch {
			case pathutil.IsExist(pathutil.JoinGoPath(filename)):
				filename = pathutil.JoinGoPath(filename)

			// filename has not directory path
			case filepath.Dir(filename) == ".":
				filename = filepath.Join(cwd, filename)

			// not contains '#' package title in errror
			case strings.HasPrefix(filename, cwd):
				sep = cwd
				filename = strings.TrimPrefix(filename, sep+string(filepath.Separator))

			// filename is like "github.com/foo/bar.go"
			case strings.HasPrefix(filename, pathutil.TrimGoPath(cwd)):
				sep = pathutil.TrimGoPath(cwd) + string(filepath.Separator)
				filename = strings.TrimPrefix(filename, sep)
			}
		case "gb":
			// gb compiler error messages is relative filename path of project root dir
			if !filepath.IsAbs(filename) {
				filename = filepath.Join(buildContext.ProjectRoot, "src", filename)
			}
		default:
			return nil, errors.New("unknown compiler tool")
		}

		// Finally, try to convert the relative path from cwd
		filename = pathutil.Rel(cwd, filename)
		if ignoreDirs != nil {
			if contains(filename, ignoreDirs) {
				continue
			}
		}

		// line is necessary for error messages
		line, err := strconv.Atoi(string(m[3]))
		if err != nil {
			return nil, err
		}

		// Ignore err because fail strconv.Atoi will assign 0 to col
		col, _ := strconv.Atoi(string(m[4]))

		errlist = append(errlist, &nvim.QuickfixError{
			FileName: filename,
			LNum:     line,
			Col:      col,
			Text:     string(bytes.TrimSpace(m[5])),
		})
	}

	return errlist, nil
}

func contains(s string, substr []string) bool {
	for _, str := range substr {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}
