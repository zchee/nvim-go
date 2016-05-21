// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	home   = os.Getenv("HOME")
	gopath = os.Getenv("GOPATH")

	goHome = filepath.Join(gopath, "/src/github.com/zchee")
	gbHome = filepath.Join(home, "/src/github.com/zchee/nvim-go")
)

func TestBuildContext(t *testing.T) {
	tests := []struct {
		// Receiver fields.
		rTool    string
		rContext build.Context
		// Parameters.
		p string
		// Expected results.
		want  string
		want1 string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		ctxt := &Build{
			Tool:    tt.rTool,
			Context: tt.rContext,
		}
		got, got1 := ctxt.buildContext(tt.p)
		if got != tt.want {
			t.Errorf("Build.buildContext(%v) got = %v, want %v", tt.p, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("Build.buildContext(%v) got1 = %v, want %v", tt.p, got1, tt.want1)
		}
	}
}

func TestIsGb(t *testing.T) {
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
			rTool:    "go",
			rContext: build.Default,
			p:        filepath.Join(goHome, "go-sandbox"), // On the  $GOPATH package
			want:     goHome + "go-sandbox",
			want1:    false,
		},
		{
			rTool:    "go",
			rContext: build.Default,
			p:        filepath.Join(goHome, "go-sandbox/llvm"), // On the  $GOPATH package
			want:     goHome + "go-sandbox",
			want1:    false,
		},
		{
			rTool:    "gb",
			rContext: build.Default,
			p:        filepath.Join(home, "/src/github.com/zchee/nvim-go"), //gb procject root directory
			want:     gbHome,
			want1:    true,
		},
		{
			rTool:    "gb",
			rContext: build.Default,
			p:        filepath.Join(home, "/src/github.com/zchee/nvim-go/src/nvim-go"), // gb source root directory
			want:     gbHome,
			want1:    true,
		},
		{
			rTool:    "gb",
			rContext: build.Default,
			p:        filepath.Join(home, "/src/github.com/zchee/nvim-go/src/nvim-go/commands"), // commands/ directory
			want:     gbHome,
			want1:    true,
		},
		{
			rTool:    "gb",
			rContext: build.Default,
			p:        filepath.Join(home, "/src/github.com/zchee/nvim-go/src/nvim-go/internel/guru"), // commands/ directory
			want:     gbHome,
			want1:    true,
		},
	}
	for _, tt := range tests {
		ctxt := &Build{
			Tool:    tt.rTool,
			Context: tt.rContext,
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

func TestSetContext(t *testing.T) {
	tests := []struct {
		// Receiver fields.
		rTool    string
		rContext build.Context
		// Parameters.
		p string
		// Expected results.
		want func()
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		c := &Build{
			Tool:    tt.rTool,
			Context: tt.rContext,
		}
		if got := c.SetContext(tt.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Build.SetContext(%v) = %v, want %v", tt.p, got, tt.want)
		}
	}
}
