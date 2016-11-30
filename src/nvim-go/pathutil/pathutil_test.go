// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"os"
	"path/filepath"
	"runtime"
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
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}

func TestJoinGoPath(t *testing.T) {
	os.Setenv("GOPATH", testGoPath)

	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "astdump",
			args: args{p: "astdump"},
			want: filepath.Join(testGoPath, "src", "astdump"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.JoinGoPath(tt.args.p); got != tt.want {
				t.Errorf("JoinGoPath(%v) = %v, want %v", tt.args.p, got, tt.want)
			}
		})
	}
}

func TestShortFilePath(t *testing.T) {
	type args struct {
		p   string
		cwd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "filename only",
			args: args{
				p:   filepath.Join(testCwd, "nvim-go/pathutil/pathutil_test.go"),
				cwd: filepath.Join(testCwd, "nvim-go/pathutil"),
			},
			want: "./pathutil_test.go",
		},
		{
			name: "with directory",
			args: args{
				p:   filepath.Join(testCwd, "nvim-go/pathutil/pathutil_test.go"),
				cwd: filepath.Join(testCwd, "nvim-go"),
			},
			want: "./pathutil/pathutil_test.go",
		},
		{
			name: "not shorten",
			args: args{
				p:   filepath.Join(testCwd, "nvim-go/pathutil/pathutil_test.go"),
				cwd: filepath.Join(testCwd, "nvim-go/commands"),
			},
			want: filepath.Join(testCwd, "nvim-go/pathutil/pathutil_test.go"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.ShortFilePath(tt.args.p, tt.args.cwd); got != tt.want {
				t.Errorf("ShortFilePath(%v, %v) = %v, want %v", tt.args.p, tt.args.cwd, got, tt.want)
			}
		})
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
			name: "own filepath and directory",
			args: args{
				f:   filepath.Join(testCwd, "pathutil_test.go"),
				cwd: testCwd,
			},
			want: "pathutil_test.go",
		},
		{
			name: "own filepath and project root",
			args: args{
				f:   filepath.Join(testCwd, "pathutil_test.go"),
				cwd: projectRoot,
			},
			want: "src/nvim-go/pathutil/pathutil_test.go",
		},
		{
			name: "Use different directory",
			args: args{
				f:   filepath.Join(testCwd, "pathutil_test.go"),
				cwd: filepath.Join(testCwd, "../commands"),
			},
			want: "../pathutil/pathutil_test.go",
		},
		{
			name: "Fail the filepath.Rel()",
			args: args{
				f:   filepath.Join(testCwd, "pathutil_test.go"),
				cwd: filepath.Join("foo", "bar", "baz"),
			},
			want: filepath.Join(testCwd, "pathutil_test.go"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.Rel(tt.args.cwd, tt.args.f); got != tt.want {
				t.Errorf("Rel(%v, %v) = %v, want %v", tt.args.f, tt.args.cwd, got, tt.want)
			}
		})
	}
}

func TestExpandGoRoot(t *testing.T) {
	goroot := runtime.GOROOT()

	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "exist $GOROOT",
			args: args{p: "$GOROOT/src/go/ast/ast.go"},
			want: filepath.Join(goroot, "src/go/ast/ast.go"),
		},
		{
			name: "not exist $GOROOT",
			args: args{p: "src/go/ast/ast.go"},
			want: "src/go/ast/ast.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.ExpandGoRoot(tt.args.p); got != tt.want {
				t.Errorf("ExpandGoRoot(%v) = %v, want %v", tt.args.p, got, tt.want)
			}
		})
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
			name: "true (own parent directory)",
			args: args{filename: testCwd},
			want: true,
		},
		{
			name: "false (own file path)",
			args: args{filename: filepath.Join(testCwd, "pathutil_test.go")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.IsDir(tt.args.filename); got != tt.want {
				t.Errorf("IsDir(%v) = %v, want %v", tt.args.filename, got, tt.want)
			}
		})
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
			name: "exist (own file)",
			args: args{filename: "./pathutil_test.go"},
			want: true,
		},
		{
			name: "not exist",
			args: args{filename: "./not_exist.go"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.IsExist(tt.args.filename); got != tt.want {
				t.Errorf("IsExist(%v) = %v, want %v", tt.args.filename, got, tt.want)
			}
		})
	}
}

func TestIsGoFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "go file",
			args: args{filename: "pathutil.go"},
			want: true,
		},
		{
			name: "not go file",
			args: args{filename: "test.c"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathutil.IsGoFile(tt.args.filename); got != tt.want {
				t.Errorf("IsGoFile(%v) = %v, want %v", tt.args.filename, got, tt.want)
			}
		})
	}
}
