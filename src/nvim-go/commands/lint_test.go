package commands

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/context"

	vim "github.com/neovim/go-client/nvim"
)

var testLintDir = filepath.Join(testGoPath, "src", "lint")

func TestCommands_Lint(t *testing.T) {
	type fields struct {
		v    *vim.Nvim
		p    *vim.Pipeline
		ctxt *context.Context
	}
	type args struct {
		args []string
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*vim.QuickfixError
		wantErr bool
	}{
		{
			name: "no suggest(go)",
			fields: fields{
				v: testVim(t, filepath.Join(testLintDir, "blank-import-main.go")),
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
				v: testVim(t, filepath.Join(testLintDir, "make.go")),
			},
			args: args{
				args: []string{"%"},
				file: filepath.Join(testLintDir, "make.go"),
			},
			want: []*vim.QuickfixError{&vim.QuickfixError{
				FileName: "make.go",
				LNum:     14,
				Col:      2,
				Text:     "can probably use \"var x []T\" instead",
			}, &vim.QuickfixError{
				FileName: "make.go",
				LNum:     15,
				Col:      2,
				Text:     "can probably use \"var y []http.Request\" instead",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ctxt := context.NewContext()
		c := NewCommands(tt.fields.v, ctxt)
		c.v.SetCurrentDirectory(filepath.Dir(tt.args.file))

		got, err := c.Lint(tt.args.args, tt.args.file)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Commands.Lint(%v, %v) error = %v, wantErr %v", tt.name, tt.args.args, tt.args.file, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Logf("%+v\n%+v", got[0], got[1])
			t.Errorf("%q. Commands.Lint(%v, %v) = %v, want %v", tt.name, tt.args.args, tt.args.file, got, tt.want)
		}
	}
}

func TestCommands_cmdLintComplete(t *testing.T) {
	type fields struct {
		v    *vim.Nvim
		p    *vim.Pipeline
		ctxt *context.Context
	}
	type args struct {
		a   *vim.CommandCompletionArgs
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
				v: testVim(t, filepath.Join(testLintDir, "make.go")),
			},
			args: args{
				a:   new(vim.CommandCompletionArgs),
				cwd: testLintDir,
			},
			wantFilelist: []string{"blank-import-main.go", "make.go", "time.go"},
		},
		{
			name: "lint dir - 'ma' (go)",
			fields: fields{
				v: testVim(t, filepath.Join(testLintDir, "make.go")),
			},
			args: args{
				a: &vim.CommandCompletionArgs{
					ArgLead: "ma",
				},
				cwd: testLintDir,
			},
			wantFilelist: []string{"make.go"},
		},
		{
			name: "astdump (go)",
			fields: fields{
				v: testVim(t, astdumpMain),
			},
			args: args{
				a:   new(vim.CommandCompletionArgs),
				cwd: astdump,
			},
			wantFilelist: []string{"astdump.go"},
		},
		{
			name: "gsftp (gb)",
			fields: fields{
				v: testVim(t, gsftpMain),
			},
			args: args{
				a:   new(vim.CommandCompletionArgs),
				cwd: gsftp,
			},
			wantFilelist: []string{"main.go"},
		},
	}
	for _, tt := range tests {
		tt.fields.ctxt = context.NewContext()
		c := NewCommands(tt.fields.v, tt.fields.ctxt)

		gotFilelist, err := c.cmdLintComplete(tt.args.a, tt.args.cwd)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Commands.cmdLintComplete(%v, %v) error = %v, wantErr %v", tt.name, tt.args.a, tt.args.cwd, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotFilelist, tt.wantFilelist) {
			t.Errorf("%q. Commands.cmdLintComplete(%v, %v) = %v, want %v", tt.name, tt.args.a, tt.args.cwd, gotFilelist, tt.wantFilelist)
		}
	}
}
