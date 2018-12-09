// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestNvim(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	type args struct {
		t    *testing.T
		file []string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []string
	}{
		{
			name:      "test",
			args:      args{t: t, file: nil},
			wantFiles: nil,
		},
		{
			name:      "Edit one file",
			args:      args{t: t, file: []string{"test.go"}},
			wantFiles: []string{filepath.Join(cwd, "test.go")},
		},
		{
			name:      "Edit two file",
			args:      args{t: t, file: []string{"test.go", "test2.go"}},
			wantFiles: []string{filepath.Join(cwd, "test.go"), filepath.Join(cwd, "test2.go")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel())

			got := TestNvim(tt.args.t, tt.args.file...)
			gotBuffers, err := got.Buffers()
			if err != nil {
				t.Log(err)
			}
			if tt.wantFiles != nil {
				for i, f := range gotBuffers {
					fname, err := got.BufferName(f)
					if err != nil {
						t.Log(err)
					}
					if tt.wantFiles[i] != fname {
						t.Errorf("tt.wantFiles[i] %s fname %s\n", tt.wantFiles[i], fname)
						t.Errorf("TestNvim(%v, %v) = %v, want %v", tt.args.t, tt.args.file, got, tt.wantFiles)
					}
				}
			}
		})
	}
}

func Test_setXDGEnv(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				dir: filepath.Join(os.TempDir(), "setXDGEnv-test"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := setXDGEnv(tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("setXDGEnv(%v) error = %v, wantErr %v", tt.args.dir, err, tt.wantErr)
			}
		})
		os.RemoveAll(tt.args.dir)
	}
}
