// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

const (
	// Buffer options
	Bufhidden    = "bufhidden"  // string
	Buflisted    = "buflisted"  // bool
	Buftype      = "buftype"    // string
	Filetype     = "filetype"   // string
	OpModifiable = "modifiable" // bool
	OpModified   = "modified"   // bool
	Swapfile     = "swapfile"   // bool

	// Buffer var
	Colorcolumn = "colorcolumn" // string

	// Window options
	List           = "list"           // bool
	Number         = "number"         // bool
	Relativenumber = "relativenumber" // bool
	Winfixheight   = "winfixheight"   // bool
)

const (
	BufhiddenDelete  = "delete"   // delete the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bdelete.
	BufhiddenHide    = "hide"     // hide the buffer (don't unload it), also when 'hidden' is not set.
	BufhiddenUnload  = "unload"   // unload the buffer, also when 'hidden' is set or using :hide.
	BufhiddenWipe    = "wipe"     // wipe out the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bwipeout.
	BuftypeAcwrite   = "acwrite"  // buffer which will always be written with BufWriteCmd autocommands.
	BuftypeHelp      = "help"     // help buffer (you are not supposed to set this manually)
	BuftypeNofile    = "nofile"   // buffer which is not related to a file and will not be written.
	BuftypeNowrite   = "nowrite"  // buffer which will not be written.
	BuftypeQuickfix  = "quickfix" // quickfix buffer, contains list of errors :cwindow or list of locations :lwindow
	BuftypeTerminal  = "terminal" // terminal buffer, this is set automatically when a terminal is created. See nvim-terminal-emulator for more information.
	FiletypeAsm      = "asm"
	FiletypeC        = "c"
	FiletypeCpp      = "cpp"
	FiletypeDelve    = "delve"
	FiletypeGas      = "gas"
	FiletypeGo       = "go"
	FiletypeTerminal = "terminal"
	FiletypeAST      = "goast"
)

type Buffer struct {
	Buffer  vim.Buffer
	Window  vim.Window
	Tabpage vim.Tabpage

	Name  string
	Bufnr int
	Mode  string
	Size  int
}

func NewBuffer(name, mode string, size int) *Buffer {
	b := &Buffer{
		Name: name,
		Mode: mode,
		Size: size,
	}

	return b
}

func (b *Buffer) Open(v *vim.Vim) error {
	err := v.Command(fmt.Sprintf("silent %s %s", b.Mode, b.Name))
	if err != nil {
		return errors.Annotate(err, "buffer.Open")
	}

	p := v.NewPipeline()
	p.CurrentBuffer(&b.Buffer)
	p.CurrentWindow(&b.Window)
	p.CurrentTabpage(&b.Tabpage)

	return p.Wait()
}

func (b *Buffer) Create(v *vim.Vim, bufOption, bufVar, winOption, winVar map[string]interface{}) error {
	defer profile.Start(time.Now(), "nvim/buffer.Create")

	err := v.Command(fmt.Sprintf("silent %s %s", b.Mode, b.Name))
	if err != nil {
		return errors.Annotate(err, "buffer.Open")
	}

	p := v.NewPipeline()
	p.CurrentBuffer(&b.Buffer)
	p.CurrentWindow(&b.Window)
	p.CurrentTabpage(&b.Tabpage)
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "nvim/buffer.Create")
	}

	p.BufferNumber(b.Buffer, &b.Bufnr)

	if bufOption != nil {
		for k, op := range bufOption {
			p.SetBufferOption(b.Buffer, k, op)
		}
	}
	if bufVar != nil {
		for k, op := range bufVar {
			p.SetBufferVar(b.Buffer, k, op, nil)
		}
	}
	if winOption != nil {
		for k, op := range winOption {
			p.SetWindowOption(b.Window, k, op)
		}
	}

	return p.Wait()
}

func (b *Buffer) UpdateSyntax(v *vim.Vim, syntax string) {
	if b.Name != "" {
		v.SetBufferName(b.Buffer, b.Name)
	}
	if syntax == "" {
		var filetype interface{}
		v.BufferOption(b.Buffer, "filetype", &filetype)
		syntax = fmt.Sprintf("%s", filetype)
	}
	v.Command(fmt.Sprintf("runtime! syntax/%s.vim", syntax))
}

const (
	Map             = "map"
	MapNormal       = "nmap"
	MapVisualSelect = "vmap"
	MapSelect       = "smap"
	MapVisual       = "xmap"
	MapOperator     = "omap"
	MapInsert       = "imap"
	MapCommandLine  = "cmap"
	MapTerminal     = "tmap"

	Noremap             = "noremap"
	NoremapNormal       = "nnoremap"
	NoremapVisualSelect = "vnoremap"
	NoremapSelect       = "snoremap"
	NoremapVisual       = "xnoremap"
	NoremapOperator     = "onoremap"
	NoremapInsert       = "inoremap"
	NoremapCommandLine  = "cnoremap"
	NoremapTerminal     = "tnoremap"
)

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

func ToByteSlice(v *vim.Vim, byt [][]byte) []byte {
	return bytes.Join(byt, []byte{'\n'})
}

func ToBufferLines(v *vim.Vim, byt []byte) [][]byte {
	return bytes.Split(byt, []byte{'\n'})
}

// ByteOffset calculation of byte offset the current cursor position.
func ByteOffset(v *vim.Vim, b vim.Buffer, w vim.Window) (int, error) {
	cursor, _ := v.WindowCursor(w)
	byteBuf, _ := v.BufferLines(b, 0, -1, true)

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

// ByteOffsetPipe calculation of byte offset the current cursor position use vim.Pipeline.
func ByteOffsetPipe(p *vim.Pipeline, b vim.Buffer, w vim.Window) (int, error) {
	var cursor [2]int
	p.WindowCursor(w, &cursor)

	var byteBuf [][]byte
	p.BufferLines(b, 0, -1, true, &byteBuf)

	if err := p.Wait(); err != nil {
		return 0, errors.Annotate(err, "nvim/buffer.ByteOffsetPipe")
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

func Contains(v *vim.Vim, b vim.Buffer) bool {
	bufs, _ := v.Buffers()
	for _, buf := range bufs {
		if buf == b {
			return true
		}
	}
	return false
}

func Modifiable(v *vim.Vim, b vim.Buffer) func() {
	v.SetBufferOption(b, OpModifiable, true)

	return func() {
		v.SetBufferOption(b, OpModifiable, false)
	}
}
