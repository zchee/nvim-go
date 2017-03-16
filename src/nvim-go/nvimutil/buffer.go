// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

const pkgBuffer = "nvim.buffer"

type BufferName string

// Buffer represents a Neovim buffer.
type Buffer struct {
	v *nvim.Nvim
	b *nvim.Batch

	buffer   nvim.Buffer
	Name     string
	Filetype string
	Bufnr    int
	Mode     string
	Height   int
	Width    int
	Data     []byte

	WindowContext
	TabpageContext
}

// NewBuffer return the new Buf instance with goroutine pipeline.
func NewBuffer(v *nvim.Nvim) *Buffer {
	return &Buffer{
		v: v,
		b: v.NewBatch(),
	}
}

func (b *Buffer) Buffer() nvim.Buffer {
	return b.buffer
}

// Create creates the new buffer and return the Buffer structure type.
func (b *Buffer) Create(name, filetype, mode string, option map[NvimOption]map[string]interface{}) error {
	b.Name = name
	b.Filetype = filetype
	b.Mode = mode

	err := b.v.Command(fmt.Sprintf("silent %s %s", b.Mode, b.Name))
	if err != nil {
		return errors.Wrap(err, pkgBuffer)
	}

	if err := b.GetBufferContext(); err != nil {
		return errors.Wrap(err, pkgBuffer)
	}

	if b.Height != 0 {
		b.b.SetWindowHeight(b.Window, b.Height)
	}
	if b.Width != 0 {
		b.b.SetWindowWidth(b.Window, b.Width)
	}

	b.b.BufferNumber(b.buffer, &b.Bufnr)

	if option != nil {
		if option[BufferOption] != nil {
			for k, op := range option[BufferOption] {
				b.b.SetBufferOption(b.buffer, k, op)
			}
		}
		if option[BufferVar] != nil {
			for k, op := range option[BufferVar] {
				b.b.SetBufferVar(b.buffer, k, op)
			}
		}
		if option[WindowOption] != nil {
			for k, op := range option[WindowOption] {
				b.b.SetWindowOption(b.Window, k, op)
			}
		}
		if option[WindowVar] != nil {
			for k, op := range option[WindowVar] {
				b.b.SetWindowVar(b.Window, k, op)
			}
		}
		if option[TabpageVar] != nil {
			for k, op := range option[TabpageVar] {
				b.b.SetTabpageVar(b.Tabpage, k, op)
			}
		}
	}

	if !strings.Contains(b.Name, ".") {
		b.b.Command(fmt.Sprintf("runtime! syntax/%s.vim", filetype))
	}

	return b.b.Execute()
}

func (b *Buffer) GetBufferContext() error {
	b.b.CurrentBuffer(&b.buffer)
	b.b.CurrentWindow(&b.Window)
	b.b.CurrentTabpage(&b.Tabpage)

	return b.b.Execute()
}

func (b *Buffer) BufferLines(start, end int, strict bool) {
	if b.buffer == 0 {
		b.GetBufferContext()
	}

	buf, err := b.v.BufferLines(b.buffer, start, end, strict)
	if err != nil {
		return
	}
	b.Data = ToByteSlice(buf)
}

func (b *Buffer) SetBufferLines(start, end int, strict bool, replacement []byte) error {
	if b.buffer == 0 {
		err := errors.New("Does not exist of target buffer")
		return err
	}

	b.Data = replacement

	return b.v.SetBufferLines(b.buffer, start, end, strict, ToBufferLines(replacement))
}

func (b *Buffer) SetBufferLinesAll(replacement []byte) error {
	if b.buffer == 0 {
		err := errors.New("Does not exist of target buffer")
		return err
	}

	b.Data = replacement
	b.Write(b.Data)

	return nil
}

// UpdateSyntax updates the syntax highlight of the buffer.
func (b *Buffer) UpdateSyntax(syntax string) {
	if b.Name != "" {
		b.v.SetBufferName(b.buffer, b.Name)
	}

	if syntax == "" {
		var filetype interface{}
		b.v.BufferOption(b.buffer, "filetype", &filetype)
		syntax = fmt.Sprintf("%s", filetype)
	}

	b.v.Command(fmt.Sprintf("runtime! syntax/%s.vim", syntax))
}

// SetLocalMapping sets buffer local mapping.
// 'mapping' arg: [key]{destination}
func (b *Buffer) SetLocalMapping(mode string, mapping map[string]string) error {
	if mapping != nil {
		cwin, err := b.v.CurrentWindow()
		if err != nil {
			return errors.Wrap(err, "nvim/buffer.SetMapping")
		}

		b.b.SetCurrentWindow(b.Window)
		defer b.v.SetCurrentWindow(cwin)

		for k, v := range mapping {
			b.b.Command(fmt.Sprintf("silent %s <buffer><silent>%s %s", mode, k, v))
		}
	}

	return b.b.Execute()
}

// lineCount counts the Neovim buffer line count and check whether 1 count,
// Because new(empty) buffer and one line buffer are both 1 count.
func (b *Buffer) lineCount() (int, error) {
	lineCount, err := b.v.BufferLineCount(b.buffer)
	if err != nil {
		return 0, errors.Wrap(err, pkgBuffer)
	}

	if lineCount == 1 {
		line, err := b.v.CurrentLine()
		if err != nil {
			return 0, errors.Wrap(err, pkgBuffer)
		}
		// Set 0 to lineCount if buffer is empty
		if len(line) == 0 {
			lineCount = 0
		}
	}

	return lineCount, nil
}

// Write appends the contents of p to the Neovim buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	lineCount, err := b.lineCount()
	if err != nil {
		return 0, errors.Wrap(err, pkgBuffer)
	}

	buf := bytes.NewBuffer(p)
	b.v.SetBufferLines(b.buffer, lineCount, -1, true, ToBufferLines(buf.Bytes()))

	return len(p), nil
}

// WriteString appends the contents of s to the Neovim buffer.
func (b *Buffer) WriteString(s string) error {
	lineCount, err := b.lineCount()
	if err != nil {
		return errors.Wrap(err, pkgBuffer)
	}

	buf := bytes.NewBufferString(s)

	return b.v.SetBufferLines(b.buffer, lineCount, -1, true, ToBufferLines(buf.Bytes()))
}

// Truncate discards all but the first n unread bytes from the
// Neovim buffer but continues to use the same allocated storage.
func (b *Buffer) Truncate(n int) {
	b.v.SetBufferLines(b.buffer, n, -1, true, [][]byte{})
}

// Reset resets the Neovim buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *Buffer) Reset() { b.Truncate(0) }

// ----------------------------------------------------------------------------
// Utility

// IsBufferValid wrapper of v.IsBufferValid function.
func IsBufferValid(v *nvim.Nvim, b nvim.Buffer) bool {
	res, err := v.IsBufferValid(b)
	if err != nil {
		return false
	}
	return res
}

// BufContains reports whether buffer list is within b.
func IsBufferContains(v *nvim.Nvim, b nvim.Buffer) bool {
	bufs, _ := v.Buffers()
	for _, buf := range bufs {
		if buf == b {
			return true
		}
	}
	return false
}

// BufExists reports whether buffer list is within bufnr use vim bufexists function.
func IsBufExists(v *nvim.Nvim, bufnr int) bool {
	var res interface{}
	v.Call("bufexists", &res, bufnr)

	return res.(int64) != 0
}

// IsVisible reports whether buffer list within buffer that &ft has filetype.
// Useful for Check qf, preview or any specific buffer is whether the opened.
func IsVisible(v *nvim.Nvim, filetype string) bool {
	buffers, err := v.Buffers()
	if err != nil {
		return false
	}
	for _, b := range buffers {
		var ft interface{}
		err := v.BufferOption(b, "filetype", &ft)
		if err != nil {
			return false
		}
		if f, ok := ft.(string); ok && f == filetype {
			return true
		}
	}
	return false
}

// Modifiable sets modifiable to true,
// The returned function restores modifiable to false.
func Modifiable(v *nvim.Nvim, b nvim.Buffer) func() {
	v.SetBufferOption(b, BufOptionModifiable, true)

	return func() {
		v.SetBufferOption(b, BufOptionModifiable, false)
	}
}

// ToByteSlice converts the 2D buffer byte data to sigle byte slice.
func ToByteSlice(byt [][]byte) []byte { return bytes.Join(byt, []byte{'\n'}) }

// ToBufferLines converts the byte slice to the 2D byte slice of Neovim buffer data.
func ToBufferLines(byt []byte) [][]byte { return bytes.Split(byt, []byte{'\n'}) }

// ByteOffset calculates the byte-offset of current cursor position.
func ByteOffset(v *nvim.Nvim, b nvim.Buffer, w nvim.Window) (int, error) {
	cursor, _ := v.WindowCursor(w)
	byteBuf, _ := v.BufferLines(b, 0, -1, true)

	if cursor[0] == 1 {
		return (1 + (cursor[1] - 1)), nil
	}

	var offset int
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

// ByteOffsetPipe calculates the byte-offset of current cursor position uses vim.Pipeline.
func ByteOffsetPipe(p *nvim.Batch, b nvim.Buffer, w nvim.Window) (int, error) {
	var cursor [2]int
	p.WindowCursor(w, &cursor)

	var byteBuf [][]byte
	p.BufferLines(b, 0, -1, true, &byteBuf)

	if err := p.Execute(); err != nil {
		return 0, errors.Wrap(err, "nvim/buffer.ByteOffsetPipe")
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
