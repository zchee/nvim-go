// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neovim/go-client/nvim"

	"nvim-go/context"
)

func TestSplitPos(t *testing.T) {
	var cwd, _ = os.Getwd()

	type args struct {
		pos string
		cwd string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 int
		want2 int
	}{
		{
			args: args{
				// strings.Split(s, sep string) []string
				pos: "/usr/local/go/src/strings/strings.go:287:6",
				cwd: cwd,
			},
			want:  "/usr/local/go/src/strings/strings.go",
			want1: 287,
			want2: 6,
		},
		{
			args: args{
				// testing.Errorf(format string, args ...interface{})
				pos: "/usr/local/go/src/testing/testing.go:482:18",
				cwd: cwd,
			},
			want:  "/usr/local/go/src/testing/testing.go",
			want1: 482,
			want2: 18,
		},
	}
	for _, tt := range tests {
		got, got1, got2 := SplitPos(tt.args.pos, tt.args.cwd)
		if got != tt.want {
			t.Errorf("%q. SplitPos(%v, %v) got = %v, want %v", tt.name, tt.args.pos, tt.args.cwd, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("%q. SplitPos(%v, %v) got1 = %v, want %v", tt.name, tt.args.pos, tt.args.cwd, got1, tt.want1)
		}
		if got2 != tt.want2 {
			t.Errorf("%q. SplitPos(%v, %v) got2 = %v, want %v", tt.name, tt.args.pos, tt.args.cwd, got2, tt.want2)
		}
	}
}

func TestParseError(t *testing.T) {
	var (
		cwd, _       = os.Getwd()
		gbProjectDir = filepath.Dir(cwd)
	)

	type args struct {
		errors []byte
		cwd    string
		ctxt   *context.Build
	}
	tests := []struct {
		name    string
		args    args
		want    []*nvim.QuickfixError
		wantErr bool
	}{
		{
			name: "gb build",
			args: args{
				errors: []byte(`# nvim-go/nvim
echo.go:79: syntax error: non-declaration statement outside function body`),
				cwd: cwd,
				ctxt: &context.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{&nvim.QuickfixError{
				FileName: "../src/nvim-go/nvim/echo.go",
				LNum:     79,
				Col:      0,
				Text:     "syntax error: non-declaration statement outside function body",
			}},
			wantErr: false,
		},
		{
			name: "gb build 2",
			args: args{
				errors: []byte(`# nvim-go/nvim/quickfix
locationlist.go:152: syntax error: unexpected case, expecting }
locationlist.go:160: syntax error: non-declaration statement outside function body`),
				cwd: cwd,
				ctxt: &context.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{
				&nvim.QuickfixError{
					FileName: "../src/nvim-go/nvim/quickfix/locationlist.go",
					LNum:     152,
					Col:      0,
					Text:     "syntax error: unexpected case, expecting }",
				},
				&nvim.QuickfixError{
					FileName: "../src/nvim-go/nvim/quickfix/locationlist.go",
					LNum:     160,
					Col:      0,
					Text:     "syntax error: non-declaration statement outside function body",
				},
			},
			wantErr: false,
		},
		{
			name: "gb build 3",
			args: args{
				errors: []byte(`# nvim-go/nvim/quickfix
locationlist.go:199: ParseError redeclared in this block
        previous declaration at locationlist.go:149`),
				cwd: cwd,
				ctxt: &context.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{
				&nvim.QuickfixError{
					FileName: "../src/nvim-go/nvim/quickfix/locationlist.go",
					LNum:     199,
					Col:      0,
					Text:     "ParseError redeclared in this block",
				},
			},
			wantErr: false,
		},
		{
			name: "go build(hyperkitctl)",
			args: args{
				errors: []byte(`# github.com/zchee/hyperkitctl/cmd/hyperkitctl
		cmd/hyperkitctl/test.go:26: undefined: hyperkitctl.WalkDir
		cmd/hyperkitctl/test.go:26: undefined: hyperkitctl.DatabasePath`),
				cwd: filepath.Join(os.Getenv("GOPATH"), "src/github.com/zchee/hyperkitctl"),
				ctxt: &context.Build{
					Tool: "go",
				},
			},
			want: []*nvim.QuickfixError{
				&nvim.QuickfixError{
					FileName: "cmd/hyperkitctl/test.go",
					LNum:     26,
					Col:      0,
					Text:     "undefined: hyperkitctl.WalkDir",
				},
				&nvim.QuickfixError{
					FileName: "cmd/hyperkitctl/test.go",
					LNum:     26,
					Col:      0,
					Text:     "undefined: hyperkitctl.DatabasePath",
				},
			},
			wantErr: false,
		},
		{
			name: "go build(hyperkitctl) 2",
			args: args{
				errors: []byte(`# github.com/zchee/hyperkitctl/cmd/hyperkitctl
		test.go:26: undefined: hyperkitctl.WalkDir
		test.go:26: undefined: hyperkitctl.DatabasePath`),
				cwd: filepath.Join(os.Getenv("GOPATH"), "src/github.com/zchee/hyperkitctl/cmd/hyperkitctl"),
				ctxt: &context.Build{
					Tool: "go",
				},
			},
			want: []*nvim.QuickfixError{
				&nvim.QuickfixError{
					FileName: "test.go",
					LNum:     26,
					Col:      0,
					Text:     "undefined: hyperkitctl.WalkDir",
				},
				&nvim.QuickfixError{
					FileName: "test.go",
					LNum:     26,
					Col:      0,
					Text:     "undefined: hyperkitctl.DatabasePath",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := ParseError(tt.args.errors, tt.args.cwd, tt.args.ctxt)
		// t.Logf("%+v", got[0])
		if (err != nil) != tt.wantErr {
			t.Errorf("%q.\nParseError(%v, %v, %v) error = %v, wantErr %v", tt.name, string(tt.args.errors), tt.args.cwd, tt.args.ctxt, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q.\nParseError(errors: %v,\ncwd: %v,\nctxt: %v) = \ngot[0]: %v, \nwant %v", tt.name, string(tt.args.errors), tt.args.cwd, tt.args.ctxt, got, tt.want)
			for _, g := range got {
				t.Logf(" got: %+v", g)
			}
			for _, w := range tt.want {
				t.Logf("want: %+v", w)
			}
		}
	}
}
