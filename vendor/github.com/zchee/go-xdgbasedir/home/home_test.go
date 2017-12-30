// Copyright 2017 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package home

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	testHomeDir := u.HomeDir

	tests := []struct {
		name string
		env  string
		want string
	}{
		{
			name: "default usrHome",
			env:  "",
			want: testHomeDir,
		},
		{
			name: "set different $HOME env",
			env:  filepath.FromSlash(filepath.Join("/tmp", "home")),
			want: filepath.FromSlash(filepath.Join("/tmp", "home")),
		},
		{
			name: "empty $HOME env",
			env:  "empty",
			want: testHomeDir,
		},
	}

	for _, tt := range tests {
		switch tt.env {
		case "":
			// nothing to do
		case "empty":
			os.Unsetenv("HOME")
		default:
			os.Setenv("HOME", tt.env)
		}

		t.Run(tt.name, func(t *testing.T) {
			if got := Dir(); got != tt.want {
				t.Errorf("Dir() = %v, want %v", got, tt.want)
			}
		})
	}
}
