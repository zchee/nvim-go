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

// BufferName represents a buffer name.
type BufferName string

// Buffer represents a Neovim buffer.
type Buffer struct {
	n *nvim.Nvim
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
func NewBuffer(n *nvim.Nvim) *Buffer {
	return &Buffer{
		n: n,
		b: n.NewBatch(),
	}
}

// Buffer return the current nvim.Buffer.
func (b *Buffer) Buffer() nvim.Buffer {
	return b.buffer
}

// Create creates the new buffer and return the Buffer structure type.
func (b *Buffer) Create(name, filetype, mode string, option map[NvimOption]map[string]interface{}) error {
	b.Name = name
	b.Filetype = filetype
	b.Mode = mode

	err := b.n.Command(fmt.Sprintf("silent %s %s", b.Mode, b.Name))
	if err != nil {
		return errors.WithStack(err)
	}

	if err := b.GetBufferContext(); err != nil {
		return errors.WithStack(err)
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

// GetBufferContext gets the current buffers context.
func (b *Buffer) GetBufferContext() error {
	b.b.CurrentBuffer(&b.buffer)
	b.b.CurrentWindow(&b.Window)
	b.b.CurrentTabpage(&b.Tabpage)

	return b.b.Execute()
}

// BufferLines gets the current buffer lines.
func (b *Buffer) BufferLines(start, end int, strict bool) {
	if b.buffer == 0 {
		b.GetBufferContext()
	}

	buf, err := b.n.BufferLines(b.buffer, start, end, strict)
	if err != nil {
		return
	}
	b.Data = ToByteSlice(buf)
}

// SetBufferLines sets the replacement to current buffer.
func (b *Buffer) SetBufferLines(start, end int, strict bool, replacement []byte) error {
	if b.buffer == 0 {
		return errors.New("Does not exist of target buffer")
	}

	b.Data = replacement

	return b.n.SetBufferLines(b.buffer, start, end, strict, ToBufferLines(replacement))
}

// SetBufferLinesAll wrapper of SetBufferLines with all lines.
func (b *Buffer) SetBufferLinesAll(replacement []byte) error {
	if b.buffer == 0 {
		return errors.New("Does not exist of target buffer")
	}

	b.Data = replacement
	b.Write(b.Data)

	return nil
}

// UpdateSyntax updates the syntax highlight of the buffer.
func (b *Buffer) UpdateSyntax(syntax string) {
	if b.Name != "" {
		b.n.SetBufferName(b.buffer, b.Name)
	}

	if syntax == "" {
		var filetype interface{}
		b.n.BufferOption(b.buffer, "filetype", &filetype)
		syntax = fmt.Sprintf("%s", filetype)
	}

	b.n.Command(fmt.Sprintf("runtime! syntax/%s.vim", syntax))
}

// SetLocalMapping sets buffer local mapping.
// 'mapping' arg: [key]{destination}
func (b *Buffer) SetLocalMapping(mode string, mapping map[string]string) error {
	if mapping != nil {
		cwin, err := b.n.CurrentWindow()
		if err != nil {
			return errors.WithStack(err)
		}

		b.b.SetCurrentWindow(b.Window)
		defer b.n.SetCurrentWindow(cwin)

		for k, v := range mapping {
			b.b.Command(fmt.Sprintf("silent %s <buffer><silent>%s %s", mode, k, v))
		}
	}

	return b.b.Execute()
}

// lineCount counts the Neovim buffer line count and check whether 1 count,
// Because new(empty) buffer and one line buffer are both 1 count.
func (b *Buffer) lineCount() (int, error) {
	lineCount, err := b.n.BufferLineCount(b.buffer)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if lineCount == 1 {
		line, err := b.n.CurrentLine()
		if err != nil {
			return 0, errors.WithStack(err)
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
		return 0, errors.WithStack(err)
	}

	buf := bytes.NewBuffer(p)
	b.n.SetBufferLines(b.buffer, lineCount, -1, true, ToBufferLines(buf.Bytes()))

	return len(p), nil
}

// WriteString appends the contents of s to the Neovim buffer.
func (b *Buffer) WriteString(s string) error {
	lineCount, err := b.lineCount()
	if err != nil {
		return errors.WithStack(err)
	}

	buf := bytes.NewBufferString(s)

	return b.n.SetBufferLines(b.buffer, lineCount, -1, true, ToBufferLines(buf.Bytes()))
}

// Truncate discards all but the first n unread bytes from the
// Neovim buffer but continues to use the same allocated storage.
func (b *Buffer) Truncate(n int) {
	b.n.SetBufferLines(b.buffer, n, -1, true, [][]byte{})
}

// Reset resets the Neovim buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *Buffer) Reset() { b.Truncate(0) }

// ----------------------------------------------------------------------------
// Utility

// IsBufferValid wrapper of nvim.IsBufferValid function.
func IsBufferValid(n *nvim.Nvim, b nvim.Buffer) bool {
	res, err := n.IsBufferValid(b)
	if err != nil {
		return false
	}
	return res
}

// IsBufferContains reports whether buffer list is within b.
func IsBufferContains(n *nvim.Nvim, b nvim.Buffer) bool {
	bufs, _ := n.Buffers()
	for _, buf := range bufs {
		if buf == b {
			return true
		}
	}
	return false
}

// IsBufExists reports whether buffer list is within bufnr use vim bufexists function.
func IsBufExists(n *nvim.Nvim, bufnr int) bool {
	var res interface{}
	n.Call("bufexists", &res, bufnr)

	return res.(int64) != 0
}

// IsVisible reports whether buffer list within buffer that &ft has filetype.
// Useful for Check qf, preview or any specific buffer is whether the opened.
func IsVisible(n *nvim.Nvim, filetype string) bool {
	buffers, err := n.Buffers()
	if err != nil {
		return false
	}
	for _, b := range buffers {
		var ft interface{}
		err := n.BufferOption(b, "filetype", &ft)
		if err != nil {
			return false
		}
		if f, ok := ft.(string); ok && f == filetype {
			return true
		}
	}
	return false
}

// Modifiable sets modifiable to true, The returned function restores modifiable to false.
func Modifiable(n *nvim.Nvim, b nvim.Buffer) func() {
	n.SetBufferOption(b, BufOptionModifiable, true)

	return func() {
		n.SetBufferOption(b, BufOptionModifiable, false)
	}
}

// ToByteSlice converts the 2D buffer byte data to sigle byte slice.
func ToByteSlice(byt [][]byte) []byte { return bytes.Join(byt, []byte{'\n'}) }

// ToBufferLines converts the byte slice to the 2D byte slice of Neovim buffer data.
func ToBufferLines(byt []byte) [][]byte { return bytes.Split(byt, []byte{'\n'}) }

// ByteOffset calculates the byte-offset of current cursor position.
func ByteOffset(n *nvim.Nvim, b nvim.Buffer, w nvim.Window) (int, error) {
	cursor, err := n.WindowCursor(w)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	byteBuf, err := n.BufferLines(b, 0, -1, true)
	if err != nil {
		return 0, errors.WithStack(err)
	}

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
