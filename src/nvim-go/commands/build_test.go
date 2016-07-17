package commands

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/config"
	"nvim-go/context"

	"github.com/neovim-go/vim"
)

func TestCommands_Build(t *testing.T) {
	type fields struct {
		Vim  *vim.Vim
		p    *vim.Pipeline
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
			name:   "nvim-go projectRoot",
			fields: fields{Vim: testVim(t, projectRoot)},
			args: args{
				eval: &CmdBuildEval{
					Cwd: projectRoot,
					Dir: projectRoot,
				},
			},
			wantErr: false,
		},
		{
			name:   "nvim-go cwd: projectRoot Dir: filepath.Join(projectRoot, \"src/nvim-go/commands\")",
			fields: fields{Vim: testVim(t, projectRoot)},
			args: args{
				eval: &CmdBuildEval{
					Cwd: projectRoot,
					Dir: filepath.Join(projectRoot, "src/nvim-go/commands"),
				},
			},
			wantErr: false,
		},
		{
			name:   "gsftp",
			fields: fields{Vim: testVim(t, gsftpRoot)},
			args: args{
				eval: &CmdBuildEval{
					Cwd: gsftpRoot,
					Dir: gsftpRoot,
				},
			},
			wantErr: false,
		},
		{
			name:   "correct (astdump)",
			fields: fields{Vim: testVim(t, filepath.Join(astdump, "astdump.go"))},
			args: args{
				eval: &CmdBuildEval{
					Cwd: astdump,
					Dir: astdump,
				},
			},
			wantErr: false,
		},
		{
			name:   "broken (astdump)",
			fields: fields{Vim: testVim(t, brokenMain)}, // broken file
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
		c := NewCommands(tt.fields.Vim)
		config.FmtMode = "goimports"
		config.ErrorListType = "locationlist"

		c.Build(tt.args.bang, tt.args.eval)
		if (len(c.errlist) != 0) != tt.wantErr {
			t.Errorf("%q. Commands.Build(%v, %v) wantErr %v", tt.name, tt.args.bang, tt.args.eval, tt.wantErr)
		}
	}
}

func TestCommands_compileCmd(t *testing.T) {
	type fields struct {
		Vim  *vim.Vim
		p    *vim.Pipeline
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
				Vim: testVim(t, projectRoot),
				ctxt: &context.Context{
					Build: context.BuildContext{Tool: "go"},
				},
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
				Vim: testVim(t, projectRoot),
				ctxt: &context.Context{
					Build: context.BuildContext{
						Tool:         "gb",
						GbProjectDir: projectRoot,
					},
				},
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
				Vim: testVim(t, projectRoot),
				ctxt: &context.Context{
					Build: context.BuildContext{
						Tool:         "gb",
						GbProjectDir: gsftpRoot,
					},
				},
			},
			args: args{
				dir: gsftpRoot,
			},
			want:    "gb",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		c := &Commands{
			v:    tt.fields.Vim,
			p:    tt.fields.p,
			ctxt: tt.fields.ctxt,
		}
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

func BenchmarkBuildGo(b *testing.B) {
	c := NewCommands(benchVim(b, astdumpMain))

	for i := 0; i < b.N; i++ {
		if err := c.Build(false, &CmdBuildEval{
			Cwd: astdump,
			Dir: astdump,
		}); err != nil {
			b.Errorf("BenchmarkBuildGo: %v", err)
		}
	}
}

func BenchmarkBuildGb(b *testing.B) {
	c := NewCommands(benchVim(b, gsftpMain))

	for i := 0; i < b.N; i++ {
		if err := c.Build(false, &CmdBuildEval{
			Cwd: gsftpRoot,
			Dir: gsftpRoot,
		}); err != nil {
			b.Errorf("BenchmarkBuildGb: %v", err)
		}
	}
}
