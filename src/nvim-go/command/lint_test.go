// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/ctx"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
)

var testLintDir = filepath.Join(testCwd, "../testdata", "go", "src", "lint")

func TestCommands_Lint(t *testing.T) {
	type fields struct {
		Nvim *nvim.Nvim
		ctx  *ctx.Context
	}
	type args struct {
		args []string
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*nvim.QuickfixError
		wantErr bool
	}{
		{
			name: "no suggest(go)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testLintDir, "blank-import-main.go")),
			},
			args: args{
				file: filepath.Join(testLintDir, "blank-import-main.go"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "2 suggest(go)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testLintDir, "time.go")),
			},
			args: args{
				args: []string{"%"},
				file: filepath.Join(testLintDir, "time.go"),
			},
			want: []*nvim.QuickfixError{&nvim.QuickfixError{
				FileName: "time.go",
				LNum:     11,
				Col:      5,
				Text:     "var rpcTimeoutMsec is of type *time.Duration; don't use unit-specific suffix \"Msec\"",
			}, &nvim.QuickfixError{
				FileName: "time.go",
				LNum:     13,
				Col:      5,
				Text:     "var timeoutSecs is of type time.Duration; don't use unit-specific suffix \"Secs\"",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctx.NewContext()
			c := NewCommand(tt.fields.Nvim, ctx)
			c.Nvim.SetCurrentDirectory(filepath.Dir(tt.args.file))

			got, err := c.Lint(tt.args.args, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. Commands.Lint(%v, %v) error = %v, wantErr %v", tt.name, tt.args.args, tt.args.file, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("%+v\n%+v", got[0], got[1])
				t.Errorf("%q. Commands.Lint(%v, %v) = %v, want %v", tt.name, tt.args.args, tt.args.file, got, tt.want)
			}
		})
	}
}

func TestCommands_cmdLintComplete(t *testing.T) {
	type fields struct {
		Nvim *nvim.Nvim
		ctx  *ctx.Context
	}
	type args struct {
		a   *nvim.CommandCompletionArgs
		cwd string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFilelist []string
		wantErr      bool
	}{
		{
			name: "lint dir - no args (go)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testLintDir, "make.go")),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: testLintDir,
			},
			wantFilelist: []string{"blank-import-main.go", "make.go", "time.go"},
		},
		{
			name: "lint dir - 'ma' (go)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testLintDir, "make.go")),
			},
			args: args{
				a: &nvim.CommandCompletionArgs{
					ArgLead: "ma",
				},
				cwd: testLintDir,
			},
			wantFilelist: []string{"make.go"},
		},
		{
			name: "astdump (go)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, astdumpMain),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: astdump,
			},
			wantFilelist: []string{"astdump.go"},
		},
		{
			name: "gsftp (gb)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, gsftpMain),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: gsftp,
			},
			wantFilelist: []string{"main.go"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fields.ctx = ctx.NewContext()
			c := NewCommand(tt.fields.Nvim, tt.fields.ctx)

			gotFilelist, err := c.cmdLintComplete(tt.args.a, tt.args.cwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. Commands.cmdLintComplete(%v, %v) error = %v, wantErr %v", tt.name, tt.args.a, tt.args.cwd, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFilelist, tt.wantFilelist) {
				t.Errorf("%q. Commands.cmdLintComplete(%v, %v) = %v, want %v", tt.name, tt.args.a, tt.args.cwd, gotFilelist, tt.wantFilelist)
			}
		})
	}
}
