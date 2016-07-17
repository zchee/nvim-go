package commands

import (
	"bytes"
	"context"
	"nvim-go/config"
	"reflect"
	"testing"

	"github.com/neovim-go/vim"
)

func TestCommands_Fmt(t *testing.T) {
	type fields struct {
		Vim  *vim.Vim
		p    *vim.Pipeline
		ctxt *context.Context
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
				Vim: testVim(t, astdumpMain), // correct file
			},
			args: args{
				dir: astdump,
			},
			wantErr: false,
		},
		{
			name: "broken (astdump)",
			fields: fields{
				Vim: testVim(t, brokenMain), // broken file
			},
			args: args{
				dir: broken,
			},
			wantErr: true,
		},
		{
			name: "correct (gsftp)",
			fields: fields{
				Vim: testVim(t, gsftpMain), // correct file
			},
			args: args{
				dir: gsftp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		c := NewCommands(tt.fields.Vim)
		config.FmtMode = "goimports"

		c.Fmt(tt.args.dir)
		if (len(c.errlist) != 0) != tt.wantErr {
			t.Errorf("%q. Commands.Fmt(%v), wantErr %v", tt.name, tt.args.dir, tt.wantErr)
		}
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
	v := testVim(t, "")

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
