// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"
)

func TestIsGb(t *testing.T) {
	var (
		cwd, _ = os.Getwd()
		gopath = os.Getenv("GOPATH")
		gbroot = filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	)

	tests := []struct {
		// Receiver fields.
		rTool    string
		rContext build.Context
		// Parameters.
		p string
		// Expected results.
		want  string
		want1 bool
	}{
		{
			rTool: "go",
			p:     filepath.Join(gopath, "/src/github.com/constabulary/gb"), // On the  $GOPATH package
			want:  filepath.Join(gopath, "/src/github.com/constabulary/gb"),
			want1: false,
		},
		{
			rTool: "go",
			p:     filepath.Join(gopath, "/src/github.com/constabulary/gb/cmd/gb"), // On the  $GOPATH package
			want:  filepath.Join(gopath, "/src/github.com/constabulary/gb"),
			want1: false,
		},
		{
			rTool: "gb",
			p:     gbroot, //gb procject root directory
			want:  gbroot, // nvim-go/src/nvim-go/context
			want1: true,
		},
		{
			rTool: "gb",
			p:     filepath.Join(gbroot, "/src/nvim-go"), // gb source root directory
			want:  gbroot,
			want1: true,
		},
		{
			rTool: "gb",
			p:     filepath.Join(gbroot, "/src/nvim-go/src/nvim-go/commands"), // commands directory
			want:  gbroot,
			want1: true,
		},
		{
			rTool: "gb",
			p:     filepath.Join(gbroot, "/src/nvim-go/src/nvim-go/internel/guru"), // internal directory
			want:  gbroot,
			want1: true,
		},
	}
	for _, tt := range tests {
		ctxt := &BuildContext{
			Tool: tt.rTool,
		}
		got, got1 := ctxt.isGb(tt.p) // projDir string, isGb bool
		if got1 != tt.want1 {        // Check the isGb package
			t.Errorf("Build.isGb(%v)\ngot1 = %v,\nwant %v", tt.p, got1, tt.want1)
		}
		if got1 && tt.rTool != "gb" { // If got1 == true, must be tt.rTool == "gb"
			t.Errorf("Build.isGb(%v)\ngot1 = %v, but rTool not %v", tt.p, got1, tt.rTool)
		}
		if tt.want1 && got != tt.want { // isGb == true but wrong projDir path. Ignore if want1 == false
			t.Errorf("Build.isGb(%v)\ngot = %v,\nwant %v", tt.p, got, tt.want)
		}
	}
}
