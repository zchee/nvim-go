// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"reflect"
	"testing"

	vim "github.com/neovim/go-client/nvim"
)

func TestNewBuffer(t *testing.T) {
	type args struct {
		v *vim.Nvim
	}
	tests := []struct {
		name string
		args args
		want *Buf
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := NewBuffer(tt.args.v); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewBuffer(%v) = %v, want %v", tt.name, tt.args.v, got, tt.want)
		}
	}
}

func TestBuf_Create(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		name     string
		filetype string
		mode     string
		option   map[NvimOption]map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.Create(tt.args.name, tt.args.filetype, tt.args.mode, tt.args.option); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.Create(%v, %v, %v, %v) error = %v, wantErr %v", tt.name, tt.args.name, tt.args.filetype, tt.args.mode, tt.args.option, err, tt.wantErr)
		}
	}
}

func TestBuf_GetBufferContext(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.GetBufferContext(); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.GetBufferContext() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestBuf_BufferLines(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		start  int
		end    int
		strict bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		b.BufferLines(tt.args.start, tt.args.end, tt.args.strict)
	}
}

func TestBuf_SetBufferLines(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		start       int
		end         int
		strict      bool
		replacement []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.SetBufferLines(tt.args.start, tt.args.end, tt.args.strict, tt.args.replacement); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.SetBufferLines(%v, %v, %v, %v) error = %v, wantErr %v", tt.name, tt.args.start, tt.args.end, tt.args.strict, tt.args.replacement, err, tt.wantErr)
		}
	}
}

func TestBuf_SetBufferLinesAll(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		replacement []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.SetBufferLinesAll(tt.args.replacement); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.SetBufferLinesAll(%v) error = %v, wantErr %v", tt.name, tt.args.replacement, err, tt.wantErr)
		}
	}
}

func TestBuf_UpdateSyntax(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		syntax string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		b.UpdateSyntax(tt.args.syntax)
	}
}

func TestBuf_SetLocalMapping(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		mode    string
		mapping map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.SetLocalMapping(tt.args.mode, tt.args.mapping); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.SetLocalMapping(%v, %v) error = %v, wantErr %v", tt.name, tt.args.mode, tt.args.mapping, err, tt.wantErr)
		}
	}
}

func TestBuf_lineCount(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		got, err := b.lineCount()
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.lineCount() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. Buf.lineCount() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestBuf_Write(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		got, err := b.Write(tt.args.p)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.Write(%v) error = %v, wantErr %v", tt.name, tt.args.p, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. Buf.Write(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}

func TestBuf_WriteString(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		if err := b.WriteString(tt.args.s); (err != nil) != tt.wantErr {
			t.Errorf("%q. Buf.WriteString(%v) error = %v, wantErr %v", tt.name, tt.args.s, err, tt.wantErr)
		}
	}
}

func TestBuf_Truncate(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		b.Truncate(tt.args.n)
	}
}

func TestBuf_Reset(t *testing.T) {
	type fields struct {
		v              *vim.Nvim
		p              *vim.Pipeline
		Buffer         vim.Buffer
		Name           string
		Filetype       string
		Bufnr          int
		Mode           string
		Data           []byte
		WindowContext  WindowContext
		TabpageContext TabpageContext
	}
	tests := []struct {
		name   string
		fields fields
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		b := &Buf{
			v:              tt.fields.v,
			p:              tt.fields.p,
			Buffer:         tt.fields.Buffer,
			Name:           tt.fields.Name,
			Filetype:       tt.fields.Filetype,
			Bufnr:          tt.fields.Bufnr,
			Mode:           tt.fields.Mode,
			Data:           tt.fields.Data,
			WindowContext:  tt.fields.WindowContext,
			TabpageContext: tt.fields.TabpageContext,
		}
		b.Reset()
	}
}

func TestIsBufferValid(t *testing.T) {
	type args struct {
		v *vim.Nvim
		b vim.Buffer
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := IsBufferValid(tt.args.v, tt.args.b); got != tt.want {
			t.Errorf("%q. IsBufferValid(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.b, got, tt.want)
		}
	}
}

func TestIsBufferContains(t *testing.T) {
	type args struct {
		v *vim.Nvim
		b vim.Buffer
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := IsBufferContains(tt.args.v, tt.args.b); got != tt.want {
			t.Errorf("%q. IsBufferContains(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.b, got, tt.want)
		}
	}
}

func TestIsBufExists(t *testing.T) {
	type args struct {
		v     *vim.Nvim
		bufnr int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := IsBufExists(tt.args.v, tt.args.bufnr); got != tt.want {
			t.Errorf("%q. IsBufExists(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.bufnr, got, tt.want)
		}
	}
}

func TestIsVisible(t *testing.T) {
	type args struct {
		v        *vim.Nvim
		filetype string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := IsVisible(tt.args.v, tt.args.filetype); got != tt.want {
			t.Errorf("%q. IsVisible(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.filetype, got, tt.want)
		}
	}
}

// func TestModifiable(t *testing.T) {
// 	type args struct {
// 		v *vim.Nvim
// 		b vim.Buffer
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func()
// 	}{
// 	// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		if got := Modifiable(tt.args.v, tt.args.b); !reflect.DeepEqual(got, tt.want) {
// 			t.Errorf("%q. Modifiable(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.b, got, tt.want)
// 		}
// 	}
// }

func TestToByteSlice(t *testing.T) {
	type args struct {
		byt [][]byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := ToByteSlice(tt.args.byt); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. ToByteSlice(%v) = %v, want %v", tt.name, tt.args.byt, got, tt.want)
		}
	}
}

func TestToBufferLines(t *testing.T) {
	type args struct {
		byt []byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := ToBufferLines(tt.args.byt); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. ToBufferLines(%v) = %v, want %v", tt.name, tt.args.byt, got, tt.want)
		}
	}
}

func TestByteOffset(t *testing.T) {
	type args struct {
		v *vim.Nvim
		b vim.Buffer
		w vim.Window
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := ByteOffset(tt.args.v, tt.args.b, tt.args.w)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. ByteOffset(%v, %v, %v) error = %v, wantErr %v", tt.name, tt.args.v, tt.args.b, tt.args.w, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. ByteOffset(%v, %v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.b, tt.args.w, got, tt.want)
		}
	}
}

func TestByteOffsetPipe(t *testing.T) {
	type args struct {
		p *vim.Pipeline
		b vim.Buffer
		w vim.Window
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := ByteOffsetPipe(tt.args.p, tt.args.b, tt.args.w)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. ByteOffsetPipe(%v, %v, %v) error = %v, wantErr %v", tt.name, tt.args.p, tt.args.b, tt.args.w, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. ByteOffsetPipe(%v, %v, %v) = %v, want %v", tt.name, tt.args.p, tt.args.b, tt.args.w, got, tt.want)
		}
	}
}
