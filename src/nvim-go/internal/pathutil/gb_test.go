// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"fmt"
	"go/build"
	"nvim-go/internal/pathutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIsGb(t *testing.T) {
	var (
		cwd, _ = os.Getwd()
		gopath = os.Getenv("GOPATH")
		gbroot = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(cwd))))
	)

	type args struct {
		dir string
	}
	tests := []struct {
		name  string
		tool  string
		args  args
		want  string
		want1 bool
	}{
		{
			name:  "go (On the $GOPATH package)",
			tool:  "go",
			args:  args{dir: filepath.Join(gopath, "/src/github.com/constabulary/gb")},
			want:  "",
			want1: false,
		},
		{
			name:  "go (On the $GOPATH package with cmd/gb directory)",
			tool:  "go",
			args:  args{dir: filepath.Join(gopath, "/src/github.com/constabulary/gb/cmd/gb")},
			want:  "",
			want1: false,
		},
		{
			name:  "gb (nvim-go root)",
			tool:  "gb",
			args:  args{dir: projectRoot}, // gb procject root directory
			want:  projectRoot,            // nvim-go/src/nvim-go/ctx
			want1: true,
		},
		{
			name:  "gb (gb source root directory)",
			tool:  "gb",
			args:  args{dir: filepath.Join(projectRoot, "src", "nvim-go")},
			want:  projectRoot,
			want1: true,
		},
		{
			name:  "gb (gb vendor directory)",
			tool:  "gb",
			args:  args{dir: filepath.Join(projectRoot, "src", "nvim-go", "vendor")},
			want:  projectRoot,
			want1: true,
		},
		{
			name:  "gb (On the gb vendor directory)",
			tool:  "gb",
			args:  args{dir: filepath.Join(projectRoot, "vendor", "src", "github.com", "neovim", "go-client", "nvim")},
			want:  projectRoot,
			want1: true,
		},
		{
			name:  "gb (nvim-go commands directory)",
			tool:  "gb",
			args:  args{dir: filepath.Join(projectRoot, "src", "nvim-go", "src", "nvim-go", "command")},
			want:  gbroot,
			want1: true,
		},
		{
			name:  "gb (nvim-go internal directory)",
			tool:  "gb",
			args:  args{dir: filepath.Join(gbroot, "src", "nvim-go", "src", "nvim-go", "internel", "guru")}, // internal directory
			want:  gbroot,
			want1: true,
		},
		{
			name:  "wrong path",
			tool:  "gb",
			args:  args{dir: "a/b/c"},
			want:  "",
			want1: false,
		},
		{
			name:  "GOROOT",
			tool:  "gb",
			args:  args{dir: filepath.Join(build.Default.GOROOT, "src", "go")},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		switch tt.tool {
		case "go":
			build.Default.GOPATH = gopath
		case "gb":
			build.Default.GOPATH = fmt.Sprintf("%s:%s/vendor", projectRoot, projectRoot)
		}

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, got1 := pathutil.IsGb(tt.args.dir)
			if got != tt.want {
				t.Errorf("IsGb(%v) got = %v, want %v", tt.args.dir, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsGb(%v) got1 = %v, want %v", tt.args.dir, got1, tt.want1)
			}
		})
	}
}

func TestFindGbProjectRoot(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "nvim-go root",
			args:    args{path: projectRoot},
			want:    projectRoot,
			wantErr: false,
		},
		{
			name:    "nvim-go with /src/commands",
			args:    args{path: filepath.Join(projectRoot, "src", "command")},
			want:    projectRoot,
			wantErr: false,
		},
		{
			name:    "gb vendor directory (return .../vendor)",
			args:    args{path: filepath.Join(projectRoot, "vendor", "src", "github.com", "neovim", "go-client", "nvim")},
			want:    filepath.Join(projectRoot, "vendor"),
			wantErr: false,
		},
		{
			name:    "gsftp",
			args:    args{path: filepath.Join(testGbPath, "gsftp", "src", "cmd", "gsftp")},
			want:    filepath.Join(testGbPath, "gsftp"),
			wantErr: false,
		},
		{
			name:    "empty path",
			args:    args{path: ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pathutil.FindGbProjectRoot(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindGbProjectRoot(%v) error = %v, wantErr %v", tt.args.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindGbProjectRoot(%v) = %v, want %v", tt.args.path, got, tt.want)
			}
		})
	}
}

func TestGbProjectName(t *testing.T) {
	type args struct {
		projectRoot string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nvim-go",
			args: args{projectRoot: projectRoot},
			want: "nvim-go",
		},
		{
			name: "gsftp",
			args: args{projectRoot: filepath.Join(testGbPath, "gsftp")},
			want: "gsftp",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := pathutil.GbProjectName(tt.args.projectRoot); got != tt.want {
				t.Errorf("GbProjectName(%v) = %v, want %v", tt.args.projectRoot, got, tt.want)
			}
		})
	}
}
