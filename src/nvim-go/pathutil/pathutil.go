// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/neovim/go-client/nvim"
)

var pkgPathutil = "pathutil"

// Chdir changes the vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func Chdir(v *nvim.Nvim, dir string) func() {
	var (
		m   sync.Mutex
		cwd interface{}
	)
	m.Lock()
	if err := v.Eval("getcwd()", &cwd); err != nil {
		return nil
	}
	v.SetCurrentDirectory(dir)
	return func() {
		v.SetCurrentDirectory(cwd.(string))
		m.Unlock()
	}
}

// Rel return the f relative path from cwd.
func Rel(f, cwd string) string {
	if filepath.HasPrefix(f, cwd) {
		return strings.TrimPrefix(f, cwd+string(filepath.Separator))
	}
	rel, _ := filepath.Rel(cwd, f)
	return rel
}

// ExpandGoRoot expands the "$GOROOT" include from p.
func ExpandGoRoot(p string) string {
	if strings.Index(p, "$GOROOT") != -1 {
		return strings.Replace(p, "$GOROOT", runtime.GOROOT(), 1)
	}

	return p // Not hit
}

// IsDir returns whether the filename is directory.
func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// IsExist returns whether the filename is exists.
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
