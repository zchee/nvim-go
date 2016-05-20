// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

import (
	"encoding/binary"
	"fmt"
	"nvim-go/nvim/profile"
	"time"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

const (
	// Buffer options
	Bufhidden  = "bufhidden"  // string
	Buflisted  = "buflisted"  // bool
	Buftype    = "buftype"    // string
	Filetype   = "filetype"   // string
	Modifiable = "modifiable" // bool
	Modified   = "modified"   // bool
	Swapfile   = "swapfile"   // bool

	// Window options
	List           = "list"           // bool
	Number         = "number"         // bool
	Relativenumber = "relativenumber" // bool
	Winfixheight   = "winfixheight"   // bool
)

const (
	BufhiddenDelete = "delete"   // delete the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bdelete.
	BufhiddenHide   = "hide"     // hide the buffer (don't unload it), also when 'hidden' is not set.
	BufhiddenUnload = "unload"   // unload the buffer, also when 'hidden' is set or using :hide.
	BufhiddenWipe   = "wipe"     // wipe out the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bwipeout.
	BuftypeAcwrite  = "acwrite"  // buffer which will always be written with BufWriteCmd autocommands.
	BuftypeHelp     = "help"     // help buffer (you are not supposed to set this manually)
	BuftypeNofile   = "nofile"   // buffer which is not related to a file and will not be written.
	BuftypeNowrite  = "nowrite"  // buffer which will not be written.
	BuftypeQuickfix = "quickfix" // quickfix buffer, contains list of errors :cwindow or list of locations :lwindow
	BuftypeTerminal = "terminal" // terminal buffer, this is set automatically when a terminal is created. See nvim-terminal-emulator for more information.
	FiletypeAsm     = "asm"
	FiletypeC       = "c"
	FiletypeCpp     = "cpp"
	FiletypeGas     = "gas"
	FiletypeGo      = "go"
	FiletypeDelve   = "delve"
)

type Buffer struct {
	Buffer  vim.Buffer
	Window  vim.Window
	Tabpage vim.Tabpage

	Name  string
	Bufnr interface{}
	Mode  string
	Size  int
}

func NewBuffer(name string) *Buffer {
	b := &Buffer{
		Name: name,
	}

	return b
}

var (
	Map             = "map"
	MapNormal       = "nmap"
	MapVisualSelect = "vmap"
	MapSelect       = "smap"
	MapVisual       = "xmap"
	MapOperator     = "omap"
	MapInsert       = "imap"
	MapCLI          = "cmap"
	MapTerminal     = "tmap"

	Noremap             = "noremap"
	NoremapNormal       = "nnoremap"
	NoremapVisualSelect = "vnoremap"
	NoremapSelect       = "snoremap"
	NoremapVisual       = "xnoremap"
	NoremapOperator     = "onoremap"
	NoremapInsert       = "inoremap"
	NoremapCLI          = "cnoremap"
	NoremapTerminal     = "tnoremap"
)

func (b *Buffer) Create(v *vim.Vim, bufOption, winOption map[string]interface{}) error {
	defer profile.Start(time.Now(), "nvim/buffer.Create")

	p := v.NewPipeline()
	p.Command(fmt.Sprintf("silent %s [delve] %s", b.Mode, b.Name))
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "nvim/buffer.Create")
	}

	p.CurrentBuffer(&b.Buffer)
	p.CurrentWindow(&b.Window)
	p.CurrentTabpage(&b.Tabpage)
	p.Eval("bufnr('%')", &b.Bufnr)
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "nvim/buffer.Create")
	}

	if bufOption != nil {
		for k, op := range bufOption {
			p.SetBufferOption(b.Buffer, k, op)
		}
	}
	if winOption != nil {
		for k, op := range winOption {
			p.SetWindowOption(b.Window, k, op)
		}
	}
	p.Command(fmt.Sprintf("runtime! syntax/%s.vim", bufOption[Filetype]))
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "nvim/buffer.Create")
	}

	// TODO(zchee): Why can't set p.SetBufferOption?
	// p.Call("setbufvar", nil, b.Bufnr.(int64), "&colorcolumn", "")

	return p.Wait()
}

// SetBufferMapping sets buffer local mapping.
// 'mapping' arg: [key]{destination}
func (b *Buffer) SetMapping(v *vim.Vim, mode string, mapping map[string]string) error {
	p := v.NewPipeline()

	if mapping != nil {
		cwin, err := v.CurrentWindow()
		if err != nil {
			return errors.Annotate(err, "nvim/buffer.SetMapping")
		}

		p.SetCurrentWindow(b.Window)
		defer v.SetCurrentWindow(cwin)

		for k, v := range mapping {
			p.Command(fmt.Sprintf("silent %s <buffer><silent>%s %s", mode, k, v))
		}
	}

	return p.Wait()
}

// ByteOffset calculation of byte offset the current cursor position.
func ByteOffset(p *vim.Pipeline) (int, error) {
	var (
		b vim.Buffer
		w vim.Window
	)

	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return 0, err
	}

	var cursor [2]int
	p.WindowCursor(w, &cursor)

	var byteBuf [][]byte
	p.BufferLines(b, 0, -1, false, &byteBuf)

	if err := p.Wait(); err != nil {
		return 0, err
	}

	if cursor[0] == 1 {
		return (1 + (cursor[1] - 1)), nil
	}

	offset := 0
	line := 1
	for _, buf := range byteBuf {
		if line == cursor[0] {
			offset++
			break
		}
		offset += (binary.Size(buf) + 1)
		line++
	}

	return (offset + (cursor[1] - 1)), nil
}
