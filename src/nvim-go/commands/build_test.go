package commands

import (
	"nvim-go/context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neovim-go/vim"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		// Parameters.
		v    *vim.Vim
		eval *CmdBuildEval
		// Expected results.
		wantErr bool
	}{
		{
			v: testVim(t, projectRoot),
			eval: &CmdBuildEval{
				Cwd: projectRoot,
				Dir: projectRoot,
			},
		},
		{
			v: testVim(t, projectRoot),
			eval: &CmdBuildEval{
				Cwd: projectRoot,
				Dir: filepath.Join(projectRoot, "src/nvim-go/commands"),
			},
		},
		{
			v: testVim(t, gsftpRoot),
			eval: &CmdBuildEval{
				Cwd: gsftpRoot,
				Dir: gsftpRoot,
			},
		},
		{
			v: testVim(t, filepath.Join(astdump, "astdump.go")),
			eval: &CmdBuildEval{
				Cwd: astdump,
				Dir: astdump,
			},
		},
	}
	for _, tt := range tests {
		if err := Build(tt.v, false, tt.eval); (err != nil) != tt.wantErr {
			t.Errorf("Build(%+v, %+v) error = %v, wantErr %v", tt.v, tt.eval, err, tt.wantErr)
		}
	}
}

func BenchmarkBuildGo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Build(benchVim(b, astdumpMain), false, &CmdBuildEval{
			Cwd: astdump,
			Dir: astdump,
		}); err != nil {
			b.Errorf("BenchmarkBuildGo: %v", err)
		}
	}
}

func BenchmarkBuildGb(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := Build(benchVim(b, gsftpMain), false, &CmdBuildEval{
			Cwd: gsftpRoot,
			Dir: gsftpRoot,
		}); err != nil {
			b.Errorf("BenchmarkBuildGb: %v", err)
		}
	}
}

func TestCompileCmd(t *testing.T) {
	goCompiler := []string{"go"}
	gbCompiler := []string{"gb"}

	tests := []struct {
		// Parameters.
		ctxt *context.Context
		dir  string
		// Expected results.
		want    []string
		wantErr bool
	}{
		{
			ctxt: &context.Context{
				Build: context.BuildContext{Tool: "go"},
			},
			dir:     astdump,
			want:    goCompiler,
			wantErr: false,
		},
		{
			ctxt: &context.Context{
				Build: context.BuildContext{
					Tool:         "gb",
					GbProjectDir: projectRoot,
				},
			},
			dir:     projectRoot,
			want:    gbCompiler,
			wantErr: false,
		},
		{
			ctxt: &context.Context{
				Build: context.BuildContext{
					Tool:         "gb",
					GbProjectDir: gsftpRoot,
				},
			},
			dir:  gsftpRoot,
			want: gbCompiler,
		},
	}
	for _, tt := range tests {
		os.Setenv("GOPATH", testGoPath)

		got, err := compileCmd(tt.ctxt, false, tt.dir)
		if (err != nil) != tt.wantErr {
			t.Errorf("compileCmd(%v, %v) error = %v, wantErr %v", tt.ctxt, tt.dir, err, tt.wantErr)
			continue
		}
		cmdArgs := got.Args[:1]
		if !reflect.DeepEqual(cmdArgs, tt.want) {
			t.Errorf("compileCmd\n%v\n%v\n\nActual %v\nwant1 %v", tt.ctxt, tt.dir, cmdArgs, tt.want)
		}
	}
}
