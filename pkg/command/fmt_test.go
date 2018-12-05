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
	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
	"github.com/zchee/nvim-go/pkg/internal/testutil"
)

func TestCommand_Fmt(t *testing.T) {
	config.FmtMode = "goimports"
	ctx := testutil.TestContext(t, context.Background())

	type fields struct {
		ctx   context.Context
		Nvim  *nvim.Nvim
		bctxt *buildctxt.Context
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
		// {
		// 	name: "correct (astdump)",
		// 	fields: fields{
		// 		ctx:  ctx,
		// 		Nvim: nvimutil.TestNvim(t, astdumpMain), // correct file
		// 		bctxt: &buildctxt.Context{
		// 			Build: buildctxt.Build{
		// 				Tool:        "go",
		// 				ProjectRoot: astdump,
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		dir: astdump,
		// 	},
		// 	wantErr: false,
		// },
		{
			name: "broken (astdump)",
			fields: fields{
				ctx:  ctx,
				Nvim: nvimutil.TestNvim(t, brokenMain), // broken file
				bctxt: &buildctxt.Context{
					Build: buildctxt.Build{
						Tool:        "go",
						ProjectRoot: broken,
					},
				},
			},
			args: args{
				dir: broken,
			},
			wantErr: true,
		},
		// {
		// 	name: "correct (gsftp)",
		// 	fields: fields{
		// 		ctx:  ctx,
		// 		Nvim: nvimutil.TestNvim(t, gsftpMain), // correct file
		// 		bctxt: &buildctxt.Context{
		// 			Build: buildctxt.Build{
		// 				Tool:        "gb",
		// 				ProjectRoot: gsftpRoot,
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		dir: gsftp,
		// 	},
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := tt.fields.ctx

			c := NewCommand(ctx, tt.fields.Nvim, tt.fields.bctxt)
			err := c.Fmt(ctx, tt.args.dir)
			switch e := err.(type) {
			case error:
				t.Errorf("%v. Commands.Fmt(%v), err %v wantErr %v", tt.name, tt.args.dir, e, tt.wantErr)
			case []*nvim.QuickfixError:
				if (len(e) != 0) != tt.wantErr {
					t.Errorf("%v. Commands.Fmt(%v), wantErr %v", tt.name, tt.args.dir, tt.wantErr)
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
	ctx := testutil.TestContext(t, context.Background())
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

		if err := minUpdate(ctx, v, b, in, out); err != nil {
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
