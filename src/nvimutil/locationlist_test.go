// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neovim/go-client/nvim"
	"github.com/zchee/nvim-go/src/buildctx"
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
		})
	}
}

func TestParseError(t *testing.T) {
	var (
		cwd, _       = os.Getwd()
		gbProjectDir = filepath.Dir(cwd)
		gopath       = build.Default.GOPATH
	)

	type args struct {
		errors       []byte
		cwd          string
		buildContext *buildctx.Build
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
				buildContext: &buildctx.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{{
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
				buildContext: &buildctx.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{
				{
					FileName: "../src/nvim-go/nvim/quickfix/locationlist.go",
					LNum:     152,
					Col:      0,
					Text:     "syntax error: unexpected case, expecting }",
				},
				{
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
				buildContext: &buildctx.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{
				{
					FileName: "../src/nvim-go/nvim/quickfix/locationlist.go",
					LNum:     199,
					Col:      0,
					Text:     "ParseError redeclared in this block",
				},
			},
			wantErr: false,
		},
		{
			name: "gb (relative path)",
			args: args{
				errors: []byte(`# nvim-go/pathutil
package_test.go:36: undeclared name: FindAll`),
				cwd: "/Users/zchee/go/src/github.com/zchee/nvim-go/src/nvim-go/command",
				buildContext: &buildctx.Build{
					Tool:        "gb",
					ProjectRoot: "/Users/zchee/go/src/github.com/zchee/nvim-go",
				},
			},
			want: []*nvim.QuickfixError{
				{
					FileName: "../pathutil/package_test.go",
					LNum:     36,
					Col:      0,
					Text:     "undeclared name: FindAll",
				},
			},
			wantErr: false,
		},
		{
			name: "have_want Go compiler type suggestion",
			args: args{
				errors: []byte(`# nvim-go/command/delve
delve.go:129: too many arguments in call to d.startServer
	 have (string, []string, string)
	 want (serverConfig, serverConfig)
delve.go:159: too many arguments in call to d.startServer
	 have (string, nil, string)
	 want (serverConfig, serverConfig)
server.go:31: cannot use cmd (type serverConfig) as type string in argument to exec.Command
server.go:33: cannot switch on cmd (type serverConfig) (struct containing []string cannot be compared)
server.go:34: invalid case "exec" in switch on cmd (mismatched types string and serverConfig)
server.go:36: invalid case "debug" in switch on cmd (mismatched types string and serverConfig)
server.go:37: cannot use cfg.flags (type []string) as type string in append
server.go:38: invalid case "connect" in switch on cmd (mismatched types string and serverConfig)
server.go:40: cannot use cfg.flags (type []string) as type string in append
FATAL: command "build" failed: exit status 2`),
				cwd: cwd,
				buildContext: &buildctx.Build{
					Tool:        "gb",
					ProjectRoot: gbProjectDir,
				},
			},
			want: []*nvim.QuickfixError{
				{
					FileName: "../src/nvim-go/command/delve/delve.go",
					LNum:     129,
					Col:      0,
					Text:     "too many arguments in call to d.startServer",
				},
				{
					FileName: "../src/nvim-go/command/delve/delve.go",
					LNum:     159,
					Col:      0,
					Text:     "too many arguments in call to d.startServer",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     31,
					Col:      0,
					Text:     "cannot use cmd (type serverConfig) as type string in argument to exec.Command",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     33,
					Col:      0,
					Text:     "cannot switch on cmd (type serverConfig) (struct containing []string cannot be compared)",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     34,
					Col:      0,
					Text:     "invalid case \"exec\" in switch on cmd (mismatched types string and serverConfig)",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     36,
					Col:      0,
					Text:     "invalid case \"debug\" in switch on cmd (mismatched types string and serverConfig)",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     37,
					Col:      0,
					Text:     "cannot use cfg.flags (type []string) as type string in append",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     38,
					Col:      0,
					Text:     "invalid case \"connect\" in switch on cmd (mismatched types string and serverConfig)",
				},
				{
					FileName: "../src/nvim-go/command/delve/server.go",
					LNum:     40,
					Col:      0,
					Text:     "cannot use cfg.flags (type []string) as type string in append",
				},
			},
			wantErr: false,
		},
	}

	build.Default.GOPATH = gopath

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseError(tt.args.errors, tt.args.cwd, tt.args.buildContext, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. ParseError(%v, %v, %v) error = %v, wantErr %v", tt.name, string(tt.args.errors), tt.args.cwd, tt.args.buildContext, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%v. ParseError(errors: %v,\ncwd: %v,\nbuildContext: %v) = \n", tt.name, string(tt.args.errors), tt.args.cwd, tt.args.buildContext)
				for _, got := range got {
					t.Logf("=====  got =====: %+v", got)
				}
				for _, want := range tt.want {
					t.Logf("===== want =====: %+v", want)
				}
			}
		})
	}
}
