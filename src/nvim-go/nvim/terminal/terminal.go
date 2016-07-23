// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package terminal

import (
	"fmt"
	"strings"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/pathutil"

	"github.com/juju/errors"
	"github.com/neovim-go/vim"
)

var pkgTerminal = "GoTerminal"

var bufName = "__GO_TERMINAL__"

// Terminal represents a Neovim terminal.
type Terminal struct {
	v *vim.Vim
	p *vim.Pipeline

	cmd  []string
	mode string
	// Name terminal buffer name.
	Name string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Size open the terminal window size.
	Size int

	cw vim.Window

	*nvim.Buf
}

// NewTerminal return the Neovim terminal buffer.
func NewTerminal(vim *vim.Vim, name string, command []string, mode string) *Terminal {
	return &Terminal{
		v:    vim,
		p:    vim.NewPipeline(),
		Name: name,
		cmd:  command,
		mode: mode,
	}
}

// Create creats the new Neovim terminal buffer.
func (t *Terminal) Create() (err error) {
	t.cw, err = t.v.CurrentWindow()
	if err != nil {
		return err
	}

	switch {
	case t.mode == "split":
		t.Size = int(config.TerminalHeight)
	case t.mode == "vsplit":
		t.Size = int(config.TerminalWidth)
	default:
		err := errors.Errorf("%s mode is not supported", t.mode)
		return nvim.ErrorWrap(t.v, errors.Annotate(err, pkgTerminal))
	}

	option := t.setTerminalOption()
	name := fmt.Sprintf("| terminal %s", strings.Join(t.cmd, " "))
	mode := fmt.Sprintf("%s %d%s", config.TerminalPosition, t.Size, t.mode)
	t.Buf = nvim.NewBuffer(t.v)
	t.Buf.Create(name, nvim.FiletypeTerminal, mode, option)
	t.Buf.Name = t.Name
	t.Buf.UpdateSyntax(nvim.FiletypeTerminal)

	// Get terminal buffer and windows information.
	t.p.CurrentBuffer(&t.Buffer)
	t.p.CurrentWindow(&t.Window)
	if err := t.p.Wait(); err != nil {
		return err
	}
	defer t.switchFocus()()

	// Cleanup cursor highlighting
	// TODO(zchee): Can use p.ClearBufferHighlight?
	t.p.Command("highlight TermCursor gui=NONE guifg=NONE guibg=NONE")
	t.p.Command("highlight TermCursorNC gui=NONE guifg=NONE guibg=NONE")

	// Set autoclose buffer if the current buffer is only terminal
	// TODO(zchee): convert to rpc way
	t.p.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")

	return t.p.Wait()
}

// Run runs the command in the terminal buffer.
func (t *Terminal) Run(cmd []string) error {
	if t.Dir != "" {
		defer pathutil.Chdir(t.v, t.Dir)()
	}

	if t.Buf != nil && nvim.IsBufferValid(t.v, t.Buffer) {
		defer t.switchFocus()()

		t.v.SetBufferOption(t.Buffer, nvim.BufOptionModified, false)
		t.v.Call("termopen", nil, cmd)
		t.v.SetBufferName(t.Buffer, t.Buf.Name)
	} else {
		t.Create()
	}
	// Workaround for "autocmd BufEnter term://* startinsert"
	if config.TerminalStartInsert {
		t.v.Command("stopinsert")
	}

	return nil
}

// TODO(zchee): flashing when switch the window.
func (t *Terminal) switchFocus() func() {
	t.v.SetCurrentWindow(t.Window)

	return func() {
		t.v.SetCurrentWindow(t.cw)
	}
}

func (t *Terminal) setTerminalOption() map[nvim.NvimOption]map[string]interface{} {
	option := make(map[nvim.NvimOption]map[string]interface{})
	bufoption := make(map[string]interface{})
	bufvar := make(map[string]interface{})
	windowoption := make(map[string]interface{})

	bufoption[nvim.BufOptionBufhidden] = nvim.BufhiddenDelete
	bufoption[nvim.BufOptionBuflisted] = false
	bufoption[nvim.BufOptionBuftype] = nvim.BuftypeNofile
	bufoption[nvim.BufOptionFiletype] = nvim.FiletypeTerminal
	bufoption[nvim.BufOptionModifiable] = false
	bufoption[nvim.BufOptionSwapfile] = false

	bufvar[nvim.BufVarColorcolumn] = ""

	windowoption[nvim.WinOptionList] = false
	windowoption[nvim.WinOptionNumber] = false
	windowoption[nvim.WinOptionRelativenumber] = false
	windowoption[nvim.WinOptionWinfixheight] = true

	option[nvim.BufferOption] = bufoption
	option[nvim.BufferVar] = bufvar
	option[nvim.WindowOption] = windowoption

	return option
}
