// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/zchee/nvim-go/src/pathutil"
)

func TestFindVCSRoot(t *testing.T) {
	testCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	type args struct {
		basedir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "go/nvim-go",
			args: args{basedir: filepath.Join(build.Default.GOPATH, "src", "github.com", "zchee", "nvim-go")},
			want: filepath.Join(build.Default.GOPATH, "src", "github.com", "zchee", "nvim-go"),
		},
		{
			name: "go/nvim-go/src/command",
			args: args{basedir: filepath.Join(build.Default.GOPATH, "src", "github.com", "zchee", "nvim-go", "src", "command")},
			want: filepath.Join(build.Default.GOPATH, "src", "github.com", "zchee", "nvim-go"),
		},
		{
			name: "go/cwd",
			args: args{basedir: testCwd},
			want: filepath.Join(build.Default.GOPATH, "src", "github.com", "zchee", "nvim-go"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := pathutil.FindVCSRoot(tt.args.basedir); got != tt.want {
				t.Errorf("FindVCSRoot(%v) = %v, want %v", tt.args.basedir, got, tt.want)
			}
		})
	}
}
