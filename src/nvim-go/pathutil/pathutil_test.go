// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	"github.com/neovim/go-client/nvim"
)

var (
	testCwd, _     = os.Getwd()
	projectRoot, _ = filepath.Abs(filepath.Join(testCwd, "../../../"))
	testdata       = filepath.Join(projectRoot, "src", "nvim-go", "testdata")
	testGoPath     = filepath.Join(testdata, "go")
	testGbPath     = filepath.Join(testdata, "gb")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
)

func TestChdir(t *testing.T) {
	type args struct {
		v   *nvim.Nvim
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantCwd string
	}{
		{
			name: "nvim-go (gb)",
			args: args{
				v:   nvimutil.TestNvim(t, projectRoot),
				dir: filepath.Join(projectRoot, "src", "nvim-go"),
			},
			wantCwd: filepath.Join(projectRoot, "src", "nvim-go"),
		},
	}
	for _, tt := range tests {
		defer func() {
			if testCwd != filepath.Join(projectRoot, "src/nvim-go/pathutil") || testCwd == tt.args.dir {
				t.Errorf("%q. Chdir(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.dir, testCwd, tt.wantCwd)
			}
		}()
		defer pathutil.Chdir(tt.args.v, tt.args.dir)()
		var ccwd interface{}
		tt.args.v.Eval("getcwd()", &ccwd)
		if ccwd.(string) != tt.wantCwd {
			t.Errorf("%q. Chdir(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.dir, ccwd, tt.wantCwd)
		}
	}
}

func TestRel(t *testing.T) {
	type args struct {
		f   string
		cwd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{f: filepath.Join(testCwd, "pathutil_test.go"), cwd: testCwd},
			want: "pathutil_test.go",
		},
		{
			args: args{f: filepath.Join(testCwd, "pathutil_test.go"), cwd: projectRoot},
			want: "src/nvim-go/pathutil/pathutil_test.go",
		},
	}
	for _, tt := range tests {
		if got := pathutil.Rel(tt.args.f, tt.args.cwd); got != tt.want {
			t.Errorf("%q. Rel(%v, %v) = %v, want %v", tt.name, tt.args.f, tt.args.cwd, got, tt.want)
		}
	}
}

func TestExpandGoRoot(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := pathutil.ExpandGoRoot(tt.args.p); got != tt.want {
			t.Errorf("%q. ExpandGoRoot(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}

func TestIsDir(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{filename: testCwd},
			want: true,
		},
		{
			args: args{filename: filepath.Join(testCwd, "pathutil_test.go")},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := pathutil.IsDir(tt.args.filename); got != tt.want {
			t.Errorf("%q. IsDir(%v) = %v, want %v", tt.name, tt.args.filename, got, tt.want)
		}
	}
}

func TestIsExist(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{filename: testCwd},
			want: true,
		},
		{
			args: args{filename: filepath.Join(testCwd, "pathutil_test.go")},
			want: true,
		},
		{
			args: args{filename: filepath.Join(testCwd, "not_exist.go")},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := pathutil.IsExist(tt.args.filename); got != tt.want {
			t.Errorf("%q. IsExist(%v) = %v, want %v", tt.name, tt.args.filename, got, tt.want)
		}
	}
}
