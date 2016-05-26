// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package terminal

import (
	"fmt"
	"strings"
	"sync"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/pathutil"

	"github.com/garyburd/neovim-go/vim"
)

var pkgTerminal = "GoTerminal"

var bufName = "__GO_TERMINAL__"

// Terminal configure of open the terminal.
type Terminal struct {
	v    *vim.Vim
	cmd  []string
	mode string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Width split window width for open the terminal window.
	Width int64
	// Height split window height for open the terminal window.
	Height int64
	Size   int

	cb     vim.Buffer
	cw     vim.Window
	Buffer *buffer.Buffer
}

// NewTerminal return the initialize Neovim terminal config.
func NewTerminal(vim *vim.Vim, command []string, mode string) *Terminal {
	return &Terminal{
		v:    vim,
		cmd:  command,
		mode: mode,
	}
}

func (t *Terminal) Create() error {
	p := t.v.NewPipeline()
	p.CurrentBuffer(&t.cb)
	p.CurrentWindow(&t.cw)
	if err := p.Wait(); err != nil {
		return err
	}

	t.Height = config.TerminalHeight
	t.Width = config.TerminalWidth

	switch {
	case t.Height != int64(0) && t.mode == "split":
		t.Size = int(t.Height)
	case t.Width != int64(0) && t.mode == "vsplit":
		t.Size = int(t.Width)
	case strings.Index(t.mode, "split") == -1:
		return fmt.Errorf("%s mode is not supported", t.mode)
	}

	t.Buffer = buffer.NewBuffer(fmt.Sprintf("| terminal %s", strings.Join(t.cmd, " ")), fmt.Sprintf("%s %d%s", config.TerminalPosition, t.Size, t.mode), t.Size)
	t.Buffer.Create(t.v, t.setTerminalOption("buffer"), t.setTerminalVar("buffer"), t.setTerminalOption("window"), t.setTerminalVar("window"))

	// Get terminal buffer and windows information.
	p.CurrentBuffer(&t.Buffer.Buffer)
	p.CurrentWindow(&t.Buffer.Window)
	if err := p.Wait(); err != nil {
		return err
	}

	// Cleanup cursor highlighting
	// TODO(zchee): Can use p.ClearBufferHighlight?
	p.Command("highlight TermCursor gui=NONE guifg=NONE guibg=NONE")
	p.Command("highlight TermCursorNC gui=NONE guifg=NONE guibg=NONE")

	// Cleanup autocmd for terminal buffer
	// The following autocmd is defined only in the terminal buffer local
	p.Command("autocmd! * <buffer>")
	// Set autoclose buffer if the current buffer is only terminal
	// TODO(zchee): convert to rpc way
	p.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")

	return p.Wait()
}

// Run runs the command in the terminal buffer.
func (t *Terminal) Run() error {
	if t.Dir != "" {
		defer pathutil.Chdir(t.v, t.Dir)()
	}

	if t.Buffer != nil && buffer.Contains(t.v, t.Buffer.Buffer) {
		defer t.switchFocus()()

		t.v.SetBufferOption(t.Buffer.Buffer, buffer.OpModified, false)
		t.v.Call("termopen", nil, t.cmd)
	} else {
		t.Create()
		defer t.switchFocus()()
	}
	// Workaround for "autocmd BufEnter term://* startinsert"
	if config.TerminalStartInsert {
		t.v.Command("stopinsert")
	}

	return nil
}

func (t *Terminal) Command(command string) error {
	defer t.switchFocus()()
	t.v.FeedKeys("i"+command+"\r", "n", true)
	return t.v.Command("stopinsert")
}

func (t *Terminal) setTerminalOption(scope string) map[string]interface{} {
	options := make(map[string]interface{})

	switch scope {
	case "buffer":
		options[buffer.Bufhidden] = buffer.BufhiddenDelete
		options[buffer.Buflisted] = false
		options[buffer.Buftype] = buffer.BuftypeNofile
		options[buffer.Filetype] = buffer.FiletypeTerminal
		options[buffer.OpModifiable] = false
		options[buffer.Swapfile] = false
	case "window":
		options[buffer.List] = false
		options[buffer.Number] = false
		options[buffer.Relativenumber] = false
		options[buffer.Winfixheight] = true
	}

	return options
}

func (t *Terminal) setTerminalVar(scope string) map[string]interface{} {
	vars := make(map[string]interface{})

	switch scope {
	case "buffer":
		vars[buffer.Colorcolumn] = ""
	}

	return vars
}

// TODO(zchee): flashing when switch the window.
func (t *Terminal) switchFocus() func() {
	t.v.SetCurrentWindow(t.Buffer.Window)

	return func() {
		t.v.SetCurrentWindow(t.cw)
	}
}

// chdir changes vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func chdir(v *vim.Vim, dir string) func() {
	var (
		m   sync.Mutex
		cwd interface{}
	)
	m.Lock()
	if err := v.Eval("getcwd()", &cwd); err != nil {
		nvim.Echoerr(v, "GoTerminal: %v", err)
	}
	v.ChangeDirectory(dir)
	return func() {
		v.ChangeDirectory(cwd.(string))
		m.Unlock()
	}
}
