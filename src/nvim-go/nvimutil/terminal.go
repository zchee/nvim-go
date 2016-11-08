// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"fmt"
	"strings"

	"nvim-go/config"
	"nvim-go/pathutil"

	"github.com/neovim/go-client/nvim"
)

var pkgTerminal = "GoTerminal"

var bufName = "__GO_TERMINAL__"

// Terminal represents a Neovim terminal.
type Terminal struct {
	v *nvim.Nvim
	p *nvim.Pipeline

	cmd  []string
	mode string
	// Name terminal buffer name.
	Name string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Size open the terminal window size.
	Size int

	cw nvim.Window

	*Buf
}

// NewTerminal return the Neovim terminal buffer.
func NewTerminal(vim *nvim.Nvim, name string, command []string, mode string) *Terminal {
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
		// nothing to do
	}

	option := t.setTerminalOption()
	name := fmt.Sprintf("| terminal %s", strings.Join(t.cmd, " "))
	mode := fmt.Sprintf("%s %d%s", config.TerminalPosition, t.Size, t.mode)
	t.Buf = NewBuffer(t.v)
	t.Buf.Create(name, FiletypeTerminal, mode, option)
	t.Buf.Name = t.Name
	t.Buf.UpdateSyntax(FiletypeTerminal)

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

	if t.Buf != nil && IsBufferValid(t.v, t.Buffer) {
		defer t.switchFocus()()

		t.v.SetBufferOption(t.Buffer, BufOptionModified, false)
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

func (t *Terminal) setTerminalOption() map[NvimOption]map[string]interface{} {
	option := make(map[NvimOption]map[string]interface{})
	bufoption := make(map[string]interface{})
	bufvar := make(map[string]interface{})
	windowoption := make(map[string]interface{})

	bufoption[BufOptionBufhidden] = BufhiddenDelete
	bufoption[BufOptionBuflisted] = false
	bufoption[BufOptionBuftype] = BuftypeNofile
	bufoption[BufOptionFiletype] = FiletypeTerminal
	bufoption[BufOptionModifiable] = false
	bufoption[BufOptionSwapfile] = false

	bufvar[BufVarColorcolumn] = ""

	windowoption[WinOptionList] = false
	windowoption[WinOptionNumber] = false
	windowoption[WinOptionRelativenumber] = false
	windowoption[WinOptionWinfixheight] = true

	option[BufferOption] = bufoption
	option[BufferVar] = bufvar
	option[WindowOption] = windowoption

	return option
}
