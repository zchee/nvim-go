// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/config"
	"nvim-go/context"

	vim "github.com/neovim/go-client/nvim"
)

func TestCommands_Build(t *testing.T) {
	type fields struct {
		v    *vim.Nvim
		ctxt *context.Context
	}
	type args struct {
		bang bool
		eval *CmdBuildEval
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nvim-go Dir: filepath.Join(projectRoot, \"src/nvim-go/commands\")",
			fields: fields{
				v:    testVim(t, projectRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd: projectRoot,
					Dir: filepath.Join(projectRoot, "src/nvim-go/commands"),
				},
			},
			wantErr: false,
		},
		{
			name: "gsftp",
			fields: fields{
				v:    testVim(t, gsftpRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd: gsftpRoot,
					Dir: gsftpRoot,
				},
			},
			wantErr: false,
		},
		{
			name: "correct (astdump)",
			fields: fields{
				v:    testVim(t, filepath.Join(astdump, "astdump.go")), // correct file
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd: astdump,
					Dir: astdump,
				},
			},
			wantErr: false,
		},
		{
			name: "broken (astdump)",
			fields: fields{
				v:    testVim(t, brokenMain), // broken file
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd: broken,
					Dir: broken,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		c := NewCommands(tt.fields.v, tt.fields.ctxt)
		config.ErrorListType = "locationlist"

		err := c.Build(tt.args.bang, tt.args.eval)
		if errlist, ok := err.([]*vim.QuickfixError); !ok {
			if (len(errlist) != 0) != tt.wantErr {
				t.Errorf("%q. Commands.Build(%v, %v) wantErr %v", tt.name, tt.args.bang, tt.args.eval, tt.wantErr)
			}
		}
	}
}

func BenchmarkBuildGo(b *testing.B) {
	ctxt := context.NewContext()
	c := NewCommands(benchVim(b, astdumpMain), ctxt)

	for i := 0; i < b.N; i++ {
		c.Build(false, &CmdBuildEval{
			Cwd: astdump,
			Dir: astdump,
		})
		if len(c.ctxt.Errlist) != 0 {
			b.Errorf("BenchmarkBuildGo: %v", c.ctxt.Errlist)
		}
	}
}

func BenchmarkBuildGb(b *testing.B) {
	ctxt := context.NewContext()
	c := NewCommands(benchVim(b, gsftpMain), ctxt)

	for i := 0; i < b.N; i++ {
		c.Build(false, &CmdBuildEval{
			Cwd: gsftpRoot,
			Dir: gsftpRoot,
		})
		if len(c.ctxt.Errlist) != 0 {
			b.Errorf("BenchmarkBuildGb: %v", c.ctxt.Errlist)
		}
	}
}

func TestCommands_compileCmd(t *testing.T) {
	type fields struct {
		Vim  *vim.Nvim
		ctxt *context.Context
	}
	type args struct {
		bang bool
		dir  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "astdump (go build)",
			fields: fields{
				Vim:  testVim(t, projectRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: astdump,
			},
			want:    "go",
			wantErr: false,
		},
		{
			name: "nvim-go (gb build)",
			fields: fields{
				Vim:  testVim(t, projectRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: projectRoot,
			},
			want:    "gb",
			wantErr: false,
		},
		{
			name: "gsftp (gb build)",
			fields: fields{
				Vim:  testVim(t, projectRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: gsftpRoot,
			},
			want:    "gb",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.fields.ctxt.Build.Tool = tt.want
		tt.fields.ctxt.Build.ProjectRoot = tt.args.dir
		c := NewCommands(tt.fields.Vim, tt.fields.ctxt)
		got, err := c.compileCmd(tt.args.bang, tt.args.dir)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Commands.compileCmd(%v, %v) error = %v, wantErr %v", tt.name, tt.args.bang, tt.args.dir, err, tt.wantErr)
			continue
		}
		cmdArgs := got.Args[0]
		if !reflect.DeepEqual(cmdArgs, tt.want) {
			t.Errorf("%q. Commands.compileCmd(%v, %v) = %v, want %v", tt.name, tt.args.bang, tt.args.dir, cmdArgs, tt.want)
		}
	}
}
