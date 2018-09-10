// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs_test

import (
	"fmt"
	"go/build"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/testutil"
)

func TestIsGb(t *testing.T) {
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
			name:  "go/wrong path",
			tool:  "go",
			args:  args{dir: "a/b/c"},
			want:  "",
			want1: false,
		},
		{
			name:  "gb/wrong path",
			tool:  "gb",
			args:  args{dir: "a/b/c"},
			want:  "",
			want1: false,
		},
		{
			name:  "go/package root",
			tool:  "go",
			args:  args{dir: filepath.Join(build.Default.GOPATH, "/src/github.com/constabulary/gb")},
			want:  "",
			want1: false,
		},
		{
			name:  "go/cmd directory",
			tool:  "go",
			args:  args{dir: filepath.Join(build.Default.GOPATH, "/src/github.com/constabulary/gb/cmd/gb")},
			want:  "",
			want1: false,
		},
		{
			name:  "go/build.Default.GOROOT",
			tool:  "go",
			args:  args{dir: filepath.Join(build.Default.GOROOT)},
			want:  "",
			want1: false,
		},
		{
			name:  "gb/package root",
			tool:  "gb",
			args:  args{dir: filepath.Join("../", "testdata", "gb", "gsftp")},
			want:  filepath.Join("../", "testdata", "gb", "gsftp"),
			want1: true,
		},
		{
			name:  "gb/src directory",
			tool:  "gb",
			args:  args{dir: filepath.Join("../", "testdata", "gb", "gsftp", "src")},
			want:  filepath.Join("../", "testdata", "gb", "gsftp"),
			want1: true,
		},
		{
			name:  "gb/vendor directory",
			tool:  "gb",
			args:  args{dir: filepath.Join("../", "testdata", "gb", "gsftp", "vendor")},
			want:  filepath.Join("../", "testdata", "gb", "gsftp"),
			want1: true,
		},
		{
			name:  "gb/vendor file",
			tool:  "gb",
			args:  args{dir: filepath.Join("../", "testdata", "gb", "gsftp", "vendor", "src", "github.com", "kr", "fs")},
			want:  filepath.Join("../", "testdata", "gb", "gsftp"),
			want1: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.tool {
			case "gb":
				root := filepath.Join("../", "testdata", "gb", "gsftp")
				defer testutil.SetBuildContext(t, fmt.Sprintf("%s:%s/vendor", root, root))()
			}
			got, got1 := fs.IsGb(tt.args.dir)
			if got != tt.want {
				t.Errorf("IsGb(%v) got = %v, want %v", tt.args.dir, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsGb(%v) got1 = %v, want %v", tt.args.dir, got1, tt.want1)
			}
		})
	}
}

func TestGbFindProjectRoot(t *testing.T) {
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
			name:    "gsftp root",
			args:    args{path: filepath.Join("../", "testdata", "gb", "gsftp")},
			want:    filepath.Join("../", "testdata", "gb", "gsftp"),
			wantErr: false,
		},
		{
			name:    "gsftp/src/cmd",
			args:    args{path: filepath.Join("../", "testdata", "gb", "gsftp", "src", "cmd")},
			want:    filepath.Join("../", "testdata", "gb", "gsftp"),
			wantErr: false,
		},
		{
			name:    "gsftp/src/cmd/gsftp",
			args:    args{path: filepath.Join("../", "testdata", "gb", "gsftp", "src", "cmd", "gsftp")},
			want:    filepath.Join("../", "testdata", "gb", "gsftp"),
			wantErr: false,
		},
		{
			name:    "gsftp vendor",
			args:    args{path: filepath.Join("../", "testdata", "gb", "gsftp", "vendor", "src", "github.com", "kr", "fs")},
			want:    filepath.Join("../", "testdata", "gb", "gsftp", "vendor"),
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

			got, err := fs.GbFindProjectRoot(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GbFindProjectRoot(%v) error = %v, wantErr %v", tt.args.path, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GbFindProjectRoot(%v) = %v, want %v", tt.args.path, got, tt.want)
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
			name: "gsftp",
			args: args{projectRoot: filepath.Join("../", "testdata", "gb", "gsftp")},
			want: "gsftp",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := fs.GbProjectName(tt.args.projectRoot); got != tt.want {
				t.Errorf("GbProjectName(%v) = %v, want %v", tt.args.projectRoot, got, tt.want)
			}
		})
	}
}

func TestGbPackages(t *testing.T) {
	type args struct {
		root string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "gsftp",
			args:    args{root: filepath.Join("../", "testdata", "gb", "gsftp")},
			want:    []string{"cmd"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fs.GbPackages(tt.args.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("GbPackages(%v) error = %v, wantErr %v", tt.args.root, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GbPackages(%v) = %v, want %v", tt.args.root, got, tt.want)
			}
		})
	}
}
