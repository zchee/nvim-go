// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"fmt"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neovim/go-client/nvim"

	"github.com/zchee/nvim-go/pkg/config"
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
