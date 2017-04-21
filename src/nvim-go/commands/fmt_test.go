// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"reflect"
	"testing"

	"nvim-go/config"
	"nvim-go/ctx"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
)

func TestCommands_Fmt(t *testing.T) {
	type fields struct {
		Nvim *nvim.Nvim
		ctx  *ctx.Context
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "correct (astdump)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, astdumpMain), // correct file
			},
			args: args{
				dir: astdump,
			},
			wantErr: false,
		},
		{
			name: "broken (astdump)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, brokenMain), // broken file
			},
			args: args{
				dir: broken,
			},
			wantErr: true,
		},
		{
			name: "correct (gsftp)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, gsftpMain), // correct file
			},
			args: args{
				dir: gsftp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		config.FmtMode = "goimports"

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := ctx.NewContext()
			c := NewCommand(tt.fields.Nvim, ctx)

			err := c.Fmt(tt.args.dir)
			if errlist, ok := err.([]*nvim.QuickfixError); !ok {
				if (len(errlist) != 0) != tt.wantErr {
					t.Errorf("%q. Commands.Fmt(%v), wantErr %v", tt.name, tt.args.dir, tt.wantErr)
				}
			}
		})
	}
}

var minUpdateTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{"a", "a"},
	{"a/b/c", "a/b/c"},

	{"a", "x"},
	{"a/b/c", "x/y/z"},

	{"a/b/c/d", "a/b/c/d"},
	{"b/c/d", "a/b/c/d"},
	{"a/b/c", "a/b/c/d"},
	{"a/d", "a/b/c/d"},
	{"a/b/c/d", "a/b/x/c/d"},

	{"a/b/c/d", "b/c/d"},
	{"a/b/c/d", "a/b/c"},
	{"a/b/c/d", "a/d"},

	{"b/c/d", "//b/c/d"},
	{"a/b/c", "a/b//c/d/"},
	{"a/b/c", "a/b//c/d/"},
	{"a/b/c/d", "a/b//c/d"},
	{"a/b/c/d", "a/b///c/d"},
}

func TestMinUpdate(t *testing.T) {
	v := nvimutil.TestNvim(t)

	b, err := v.CurrentBuffer()
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range minUpdateTests {
		in := bytes.Split([]byte(tt.in), []byte{'/'})
		out := bytes.Split([]byte(tt.out), []byte{'/'})

		if err := v.SetBufferLines(b, 0, -1, true, in); err != nil {
			t.Fatal(err)
		}

		if err := minUpdate(v, b, in, out); err != nil {
			t.Errorf("%q -> %q returned %v", tt.in, tt.out, err)
			continue
		}

		actual, err := v.BufferLines(b, 0, -1, true)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(actual, out) {
			t.Errorf("%q -> %q returned %v, want %v", tt.in, tt.out, actual, out)
			continue
		}
	}
}
