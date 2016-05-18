// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

import (
	"encoding/binary"
	"fmt"

	"github.com/garyburd/neovim-go/vim"
	"github.com/pkg/errors"
)

const (
	// Buffer options
	Filetype  = "filetype"
	Buftype   = "buftype"
	Bufhidden = "bufhidden"
	Buflisted = "buflisted"
	Swapfile  = "swapfile"

	// Window options
	List           = "list"
	Number         = "number"
	Relativenumber = "relativenumber"
	Winfixheight   = "winfixheight"

	FiletypeAsm     = "asm"
	FiletypeC       = "c"
	FiletypeCpp     = "cpp"
	FiletypeGo      = "go"
	BuftypeNofile   = "nofile"   // buffer which is not related to a file and will not be written.
	BuftypeNowrite  = "nowrite"  // buffer which will not be written.
	BuftypeAcwrite  = "acwrite"  // buffer which will always be written with BufWriteCmd autocommands.
	BuftypeQuickfix = "quickfix" // quickfix buffer, contains list of errors :cwindow or list of locations :lwindow
	BuftypeHelp     = "help"     // help buffer (you are not supposed to set this manually)
	BuftypeTerminal = "terminal" // terminal buffer, this is set automatically when a terminal is created. See nvim-terminal-emulator for more information.
	BufhiddenHide   = "hide"     // hide the buffer (don't unload it), also when 'hidden' is not set.
	BufhiddenUnload = "unload"   // unload the buffer, also when 'hidden' is set or using :hide.
	BufhiddenDelete = "delete"   // delete the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bdelete.
	BufhiddenWipe   = "wipe"     // wipe out the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bwipeout.
)

type Buffer struct {
	Buffer  vim.Buffer
	Window  vim.Window
	Tabpage vim.Tabpage

	Name  string
	Bufnr interface{}
	Mode  string
}

func NewBuffer(name string) *Buffer {
	b := &Buffer{
		Name: name,
	}

	return b
}

func (b *Buffer) Create(v *vim.Vim, bufOption, winOption map[string]interface{}) error {
	p := v.NewPipeline()
	p.Command(fmt.Sprintf("silent %s [delve] %s", b.Mode, b.Name))
	if err := p.Wait(); err != nil {
		return errors.Wrap(err, "Delve")
	}

	p.CurrentBuffer(&b.Buffer)
	p.CurrentWindow(&b.Window)
	if err := p.Wait(); err != nil {
		return errors.Wrap(err, "Delve")
	}

	p.Eval("bufnr('%')", b.Bufnr)
	for k, v := range bufOption {
		p.SetBufferOption(b.Buffer, k, v)
	}
	for k, v := range winOption {
		p.SetWindowOption(b.Window, k, v)
	}
	if err := p.Wait(); err != nil {
		return errors.Wrap(err, "Delve")
	}

	// TODO(zchee): Why can't set p.SetBufferOption?
	// p.Call("setbufvar", nil, b.bufnr.(int64), "&colorcolumn", "")

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
