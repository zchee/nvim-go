// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"go/build"
	"path/filepath"
	"testing"

	"github.com/zchee/nvim-go/src/pathutil"
)

func TestPackagePath(t *testing.T) {
	oldGoPath := build.Default.GOPATH
	testGoPath := filepath.Join("testdata", "go")
	build.Default.GOPATH = testGoPath
	defer func() {
		build.Default.GOPATH = oldGoPath
	}()

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "package main (main.go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "testmain")},
			want:    filepath.Join(testGoPath, "src", "foo.org", "testmain"),
			wantErr: false,
		},
		{
			name:    "package foo (exists go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo")},
			want:    filepath.Join(testGoPath, "src", "foo.org", "foo"),
			wantErr: false,
		},
		{
			name:    "not exists go file(use parent dir)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar")},
			want:    filepath.Join(testGoPath, "src", "foo.org", "foo"),
			wantErr: false,
		},
		{
			name:    "package baz (parent dir is no go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz")},
			want:    filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz"),
			wantErr: false,
		},
		{
			name:    "package qux (parent dir is package)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz", "qux")},
			want:    filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz", "qux"),
			wantErr: false,
		},
		{
			name:    "no such file or directory",
			args:    args{dir: filepath.Join("nosuch", "src", "foo.org", "notexists")},
			want:    "",
			wantErr: true,
		},
		{
			name:    "GOPATH directory",
			args:    args{dir: testGoPath},
			want:    "",
			wantErr: true,
		},
		{
			name:    "GOROOT directory",
			args:    args{dir: build.Default.GOROOT},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := pathutil.PackagePath(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackagePath(%v) error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PackagePath(%v) = got: %v, want %v", tt.args.dir, got, tt.want)
			}
		})
	}
}

func TestPackageID(t *testing.T) {
	oldGoPath := build.Default.GOPATH
	testGoPath := filepath.Join("testdata", "go")
	build.Default.GOPATH = testGoPath
	defer func() {
		build.Default.GOPATH = oldGoPath
	}()

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "package main (main.go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "testmain")},
			want:    filepath.Join("foo.org", "testmain"),
			wantErr: false,
		},
		{
			name:    "package foo (exists go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo")},
			want:    filepath.Join("foo.org", "foo"),
			wantErr: false,
		},
		{
			name:    "not exists go file(use parent dir)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar")},
			want:    filepath.Join("foo.org", "foo"),
			wantErr: false,
		},
		{
			name:    "package baz (parent dir is no go file)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz")},
			want:    filepath.Join("foo.org", "foo", "bar", "baz"),
			wantErr: false,
		},
		{
			name:    "package qux (parent dir is package)",
			args:    args{dir: filepath.Join(testGoPath, "src", "foo.org", "foo", "bar", "baz", "qux")},
			want:    filepath.Join("foo.org", "foo", "bar", "baz", "qux"),
			wantErr: false,
		},
		{
			name:    "no such file or directory",
			args:    args{dir: filepath.Join("nosuch", "src", "foo.org", "notexists")},
			want:    "",
			wantErr: true,
		},
		{
			name:    "GOPATH directory",
			args:    args{dir: testGoPath},
			want:    "",
			wantErr: true,
		},
		{
			name:    "GOROOT directory",
			args:    args{dir: build.Default.GOROOT},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := pathutil.PackageID(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackageID(%v) error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PackageID(%v) = %v, want %v", tt.args.dir, got, tt.want)
			}
		})
	}
}
