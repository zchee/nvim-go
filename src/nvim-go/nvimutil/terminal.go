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
	b *nvim.Batch

	cmd  []string
	mode string
	// Name terminal buffer name.
	Name string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Size open the terminal window size.
	Size int

	cw nvim.Window

	*Buffer
}

// NewTerminal return the Neovim terminal buffer.
func NewTerminal(vim *nvim.Nvim, name string, command []string, mode string) *Terminal {
	return &Terminal{
		v:    vim,
		b:    vim.NewBatch(),
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

	t.Buffer = NewBuffer(t.v)

	switch {
	case t.mode == "split":
		t.Size = t.getWindowSize(config.TerminalHeight, t.v.WindowHeight)
		t.Buffer.Height = t.Size
	case t.mode == "vsplit":
		t.Size = t.getWindowSize(config.TerminalWidth, t.v.WindowWidth)
		t.Buffer.Width = t.Size
	default:
		// nothing to do
	}

	option := t.setTerminalOption()
	name := fmt.Sprintf("| terminal %s", strings.Join(t.cmd, " "))
	mode := fmt.Sprintf("%s %d%s", config.TerminalPosition, t.Size, t.mode)

	t.Buffer.Create(name, FiletypeTerminal, mode, option)
	t.Buffer.Name = t.Name
	t.Buffer.UpdateSyntax(FiletypeGoTerminal)

	defer t.switchFocus()()

	// Cleanup cursor highlighting
	// TODO(zchee): Can use p.ClearBufferHighlight?
	t.b.Command("highlight TermCursor gui=NONE guifg=NONE guibg=NONE")
	t.b.Command("highlight TermCursorNC gui=NONE guifg=NONE guibg=NONE")

	// Set autoclose buffer if the current buffer is only terminal
	// TODO(zchee): convert to rpc way
	t.b.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")

	return t.b.Execute()
}

// Run runs the command in the terminal buffer.
func (t *Terminal) Run(cmd []string) error {
	if t.Dir != "" {
		defer pathutil.Chdir(t.v, t.Dir)()
	}

	if t.Buffer != nil && IsBufferValid(t.v, t.buffer) {
		defer t.switchFocus()()

		t.v.SetBufferOption(t.buffer, BufOptionModified, false)
		t.v.Call("termopen", nil, cmd)
		t.v.SetBufferName(t.buffer, t.Buffer.Name)
	} else {
		t.Create()
	}
	// Workaround for "autocmd BufEnter term://* startinsert"
	if config.TerminalStopInsert {
		t.v.Command("stopinsert")
	}

	return nil
}

// getWindowSize return the one third of window (height|width) size if cfg is 0
func (t *Terminal) getWindowSize(cfg int64, fn func(nvim.Window) (int, error)) int {
	if cfg == 0 {
		i, err := fn(t.cw)
		if err != nil {
			return 0
		}
		return i / 3
	}
	return int(cfg)
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
