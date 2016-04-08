// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/garyburd/neovim-go/vim"
)

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
	v, err := vim.StartEmbeddedVim(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n"},
		Env:  []string{},
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

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
