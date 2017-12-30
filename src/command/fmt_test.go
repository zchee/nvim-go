// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/neovim/go-client/nvim"
	"github.com/zchee/nvim-go/src/buildctx"
	"github.com/zchee/nvim-go/src/config"
	"github.com/zchee/nvim-go/src/nvimutil"
	"github.com/zchee/nvim-go/src/testutil"
)

func TestCommand_Fmt(t *testing.T) {
	ctx := testutil.TestContext(context.Background())

	type fields struct {
		ctx       context.Context
		Nvim      *nvim.Nvim
		buildctxt *buildctx.Context
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
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, astdumpMain), // correct file
				buildctxt: buildctx.NewContext(),
			},
			args: args{
				dir: astdump,
			},
			wantErr: false,
		},
		{
			name: "broken (astdump)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, brokenMain), // broken file
				buildctxt: buildctx.NewContext(),
			},
			args: args{
				dir: broken,
			},
			wantErr: true,
		},
		{
			name: "correct (gsftp)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, gsftpMain), // correct file
				buildctxt: buildctx.NewContext(),
			},
			args: args{
				dir: gsftp,
			},
			wantErr: false,
		},
	}
	config.FmtMode = "goimports"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCommand(tt.fields.ctx, tt.fields.Nvim, tt.fields.buildctxt)

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
