// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/neovim-go/vim"
)

func TestFmt(t *testing.T) {
	tests := []struct {
		// Parameters.
		v   *vim.Vim
		dir string
		// Expected results.
		wantErr bool
	}{
		{
			v:       testVim(t, astdumpMain), // correct file
			dir:     astdump,
			wantErr: false,
		},
		{
			v:       testVim(t, brokenMain), // broken file
			dir:     broken,
			wantErr: true,
		},
		{
			v:       testVim(t, gsftpMain), // correct file
			dir:     gsftp,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if stat, err := os.Stat(tt.dir); err != nil || !stat.IsDir() {
			t.Error(err)
		}
		if err := Fmt(tt.v, tt.dir); (err != nil) != tt.wantErr {
			t.Errorf("Fmt(%v, %v) error = %v, wantErr %v", tt.v, tt.dir, err, tt.wantErr)
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
