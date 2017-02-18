package commands

import (
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/context"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
)

func TestCommands_Build(t *testing.T) {
	type fields struct {
		Nvim     *nvim.Nvim
		Pipeline *nvim.Pipeline
		Batch    *nvim.Batch
		ctxt     *context.Context
	}
	type args struct {
		bang bool
		eval *CmdBuildEval
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		// want interface{}
		wantErr bool
	}{
		{
			name: "nvim-go File: filepath.Join(projectRoot, \"src/nvim-go/commands\")",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, "testdata"),
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd:  testdataPath,
					File: testdataPath,
				},
			},
			wantErr: false,
		},
		{
			name: "gsftp",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, gsftpRoot),
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd:  gsftpRoot,
					File: gsftpRoot,
				},
			},
			wantErr: false,
		},
		{
			name: "correct (astdump)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(astdump, "astdump.go")), // correct file
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd:  astdump,
					File: filepath.Join(astdump, "astdump.go"),
				},
			},
			wantErr: false,
		},
		{
			name: "broken (broken)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, brokenMain), // broken file
				ctxt: context.NewContext(),
			},
			args: args{
				eval: &CmdBuildEval{
					Cwd:  broken,
					File: broken,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCommands(tt.fields.Nvim, tt.fields.ctxt)
			// if got := c.Build(tt.args.bang, tt.args.eval); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Commands.Build(%v, %v) = %v, want %v", tt.args.bang, tt.args.eval, got, tt.want)
			// }
			err := c.Build(tt.args.bang, tt.args.eval)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v", tt.wantErr)
				return
			}
			if errlist, ok := err.([]*nvim.QuickfixError); !ok {
				if (len(errlist) != 0) != tt.wantErr {
					t.Errorf("%q. Commands.Build(%v, %v)", tt.name, tt.args.bang, tt.args.eval)
				}
			}
		})
	}
}

func BenchmarkBuildGo(b *testing.B) {
	ctxt := context.NewContext()
	c := NewCommands(benchVim(b, astdumpMain), ctxt)

	for i := 0; i < b.N; i++ {
		c.Build(false, &CmdBuildEval{
			Cwd:  astdump,
			File: astdump,
		})
		if len(c.ctx.Errlist) != 0 {
			b.Errorf("BenchmarkBuildGo: %v", c.ctx.Errlist)
		}
	}
}

func BenchmarkBuildGb(b *testing.B) {
	ctxt := context.NewContext()
	c := NewCommands(benchVim(b, gsftpMain), ctxt)

	for i := 0; i < b.N; i++ {
		c.Build(false, &CmdBuildEval{
			Cwd:  gsftpRoot,
			File: gsftpRoot,
		})
		if len(c.ctx.Errlist) != 0 {
			b.Errorf("BenchmarkBuildGb: %v", c.ctx.Errlist)
		}
	}
}

func TestCommands_compileCmd(t *testing.T) {
	gobinary, err := exec.LookPath("go")
	if err != nil {
		t.Error(err)
	}
	gbbinary, err := exec.LookPath("gb")
	if err != nil {
		t.Error(err)
	}

	type fields struct {
		Nvim     *nvim.Nvim
		Pipeline *nvim.Pipeline
		Batch    *nvim.Batch
		ctxt     *context.Context
	}
	type args struct {
		bang bool
		dir  string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantErr  bool
		testfile bool
	}{
		{
			name: "astdump (go build)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, "testdata"),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: astdump,
			},
			want:    gobinary,
			wantErr: false,
		},
		{
			name: "nvim-go (gb build)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, "testdata"),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: "testdata",
			},
			want:    gbbinary,
			wantErr: false,
		},
		{
			name: "gsftp (gb build)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, "testdata"),
				ctxt: context.NewContext(),
			},
			args: args{
				dir: gsftpRoot,
			},
			want:    gbbinary,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.fields.ctxt.Build.Tool = tt.want
		tt.fields.ctxt.Build.ProjectRoot = tt.args.dir

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCommands(tt.fields.Nvim, tt.fields.ctxt)

			got, err := c.compileCmd(tt.args.bang, tt.args.dir, tt.testfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Commands.compileCmd(%v, %v) error = %v, wantErr %v", tt.args.bang, tt.args.dir, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Args[0], tt.want) {
				t.Errorf("Commands.compileCmd(%v, %v) = %v, want %v", tt.args.bang, tt.args.dir, got, tt.want)
			}
		})
	}
}
