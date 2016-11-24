// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"go/build"
	"nvim-go/pathutil"
	"testing"
)

func TestPackagePath(t *testing.T) {
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
			name:    "astdump",
			args:    args{dir: astdump},
			want:    astdump,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		build.Default.GOPATH = testGoPath

		t.Run(tt.name, func(t *testing.T) {
			got, err := pathutil.PackagePath(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackagePath(%v) error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PackagePath(%v) = %v, want %v", tt.args.dir, got, tt.want)
			}
		})
	}
}

func TestPackageID(t *testing.T) {
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
			name:    "astdump",
			args:    args{dir: astdump},
			want:    "astdump",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		build.Default.GOPATH = testGoPath

		t.Run(tt.name, func(t *testing.T) {
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
