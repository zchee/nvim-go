// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"fmt"
	"go/build"
	"nvim-go/pathutil"
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
			name:  "go github.com/constabulary/gb",
			tool:  "go",
			args:  args{dir: filepath.Join(gopath, "/src/github.com/constabulary/gb")}, // On the $GOPATH package
			want:  "",
			want1: false,
		},
		{
			name:  "go github.com/constabulary/gb/cmd/gb",
			tool:  "go",
			args:  args{dir: filepath.Join(gopath, "/src/github.com/constabulary/gb/cmd/gb")}, // On the $GOPATH package
			want:  "",
			want1: false,
		},
		{
			name:  "gb (nvim-go root)",
			tool:  "gb",
			args:  args{dir: gbroot}, //gb procject root directory
			want:  gbroot,            // nvim-go/src/nvim-go/context
			want1: true,
		},
		{
			name:  "gb (nvim-go/src/nvim-go)",
			tool:  "gb",
			args:  args{dir: filepath.Join(gbroot, "/src/nvim-go")}, // gb source root directory
			want:  gbroot,
			want1: true,
		},
		{
			name:  "gb (nvim-go/src/nvim-go/commands)",
			tool:  "gb",
			args:  args{dir: filepath.Join(gbroot, "/src/nvim-go/src/nvim-go/commands")}, // commands directory
			want:  gbroot,
			want1: true,
		},
		{
			name:  "gb (nvim-go/src/nvim-go/commands/guru)",
			tool:  "gb",
			args:  args{dir: filepath.Join(gbroot, "/src/nvim-go/src/nvim-go/internel/guru")}, // internal directory
			want:  gbroot,
			want1: true,
		},
	}
	for _, tt := range tests {
		switch tt.tool {
		case "go":
			build.Default.GOPATH = gopath
		case "gb":
			build.Default.GOPATH = fmt.Sprintf("%s:%s/vendor", projectRoot, projectRoot)
		}

		t.Run(tt.name, func(t *testing.T) {
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
			args:    args{path: filepath.Join(projectRoot, "src", "commands")},
			want:    projectRoot,
			wantErr: false,
		},
		{
			name:    "gsftp",
			args:    args{path: filepath.Join(testGbPath, "gsftp", "src", "cmd", "gsftp")},
			want:    filepath.Join(testGbPath, "gsftp"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.GbProjectName(tt.args.projectRoot); got != tt.want {
				t.Errorf("GbProjectName(%v) = %v, want %v", tt.args.projectRoot, got, tt.want)
			}
		})
	}
}
