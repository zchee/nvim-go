// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"fmt"
	"strings"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/pathutil"
)

var terminalBufferName = "__GO_TERMINAL__"

// Terminal represents a Neovim terminal.
type Terminal struct {
	cmd  []string
	mode string
	// Name terminal buffer name.
	Name string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Size open the terminal window size.
	Size int

	Nvim  *nvim.Nvim
	Batch *nvim.Batch
	cw    nvim.Window

	*Buffer
}

// NewTerminal return the Neovim terminal buffer.
func NewTerminal(n *nvim.Nvim, name string, command []string, mode string) *Terminal {
	if name == "" {
		name = terminalBufferName
	}
	return &Terminal{
		cmd:   command,
		mode:  mode,
		Name:  name,
		Nvim:  n,
		Batch: n.NewBatch(),
	}
}

// Create creats the new Neovim terminal buffer.
func (t *Terminal) Create() (err error) {
	t.cw, err = t.Nvim.CurrentWindow()
	if err != nil {
		return err
	}

	t.Buffer = NewBuffer(t.Nvim)
	t.Buffer.Filetype = FiletypeGoTerminal

	switch {
	case t.mode == "split":
		t.Size = t.getSplitWindowSize(config.TerminalHeight, t.Nvim.WindowHeight)
		t.Buffer.Height = t.Size
	case t.mode == "vsplit":
		t.Size = t.getSplitWindowSize(config.TerminalWidth, t.Nvim.WindowWidth)
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
	t.Batch.Command("highlight TermCursor gui=NONE guifg=NONE guibg=NONE")
	t.Batch.Command("highlight TermCursorNC gui=NONE guifg=NONE guibg=NONE")

	// Set autoclose buffer if the current buffer is only terminal
	// TODO(zchee): convert to rpc way
	t.Batch.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")

	return t.Batch.Execute()
}

// Run runs the command in the terminal buffer.
func (t *Terminal) Run(cmd []string) error {
	if t.Dir != "" {
		defer pathutil.Chdir(t.Nvim, t.Dir)()
	}

	if t.Buffer != nil && IsBufferValid(t.Nvim, t.buffer) {
		defer t.switchFocus()()

		t.Nvim.SetBufferOption(t.buffer, BufOptionModified, false)
		t.Nvim.Call("termopen", nil, cmd)
		t.Nvim.SetBufferName(t.buffer, t.Buffer.Name)
	} else {
		t.Create()
	}
	// Workaround for "autocmd BufEnter term://* startinsert"
	if config.TerminalStopInsert {
		t.Nvim.Command("stopinsert")
	}

	lines, err := t.Nvim.BufferLineCount(t.buffer)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(t.Nvim.SetWindowCursor(t.Window, [2]int{lines, 0}))
}

// getSplitWindowSize return the one third of window (height|width) size if cfg is 0
func (t *Terminal) getSplitWindowSize(cfg int64, f func(nvim.Window) (int, error)) int {
	if cfg == 0 {
		i, err := f(t.cw)
		if err != nil {
			return 0
		}
		return i / 3
	}
	return int(cfg)
}

// TODO(zchee): flashing when switch the window.
func (t *Terminal) switchFocus() func() {
	t.Nvim.SetCurrentWindow(t.Window)

	return func() {
		t.Nvim.SetCurrentWindow(t.cw)
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
