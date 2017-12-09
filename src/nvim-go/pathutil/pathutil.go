// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/neovim/go-client/nvim"
)

// Chdir changes the vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func Chdir(v *nvim.Nvim, dir string) func() {
	var m sync.Mutex
	var cwd interface{}

	m.Lock()
	v.Eval("getcwd()", &cwd)
	v.SetCurrentDirectory(dir)
	return func() {
		v.SetCurrentDirectory(cwd.(string))
		m.Unlock()
	}
}

// JoinGoPath joins the $GOPATH + "src" to p
func JoinGoPath(p string) string {
	return filepath.Join(build.Default.GOPATH, "src", p)
}

// TrimGoPath trims the GOPATH and {bin,pkg,src}, basically for the converts
// the package ID
func TrimGoPath(p string) string {
	// Separate trim work for p equal GOPATH
	p = strings.TrimPrefix(p, build.Default.GOPATH+string(filepath.Separator))

	if len(p) >= 4 {
		switch p[:3] {
		case "bin", "pkg", "src":
			return filepath.Clean(p[4:])
		}
	}

	return p
}

// ExpandGoRoot expands the "$GOROOT" include from p.
func ExpandGoRoot(p string) string {
	if strings.Index(p, "$GOROOT") != -1 {
		return strings.Replace(p, "$GOROOT", runtime.GOROOT(), 1)
	}

	return p // Not hit
}

// ShortFilePath return the simply trim cwd into p.
func ShortFilePath(p, cwd string) string {
	return strings.Replace(p, cwd, ".", 1)
}

// Rel wrapper of filepath.Rel function that return only one variable.
func Rel(cwd, f string) string {
	rel, err := filepath.Rel(cwd, f)
	if err != nil {
		return f
	}
	return rel
}

// ToWildcard returns the path with wildcard(...) suffix.
func ToWildcard(path string) string {
	return path + string(filepath.Separator) + "..."
}

func Create(filename string) error {
	if IsNotExist(filename) {
		if _, err := os.Create(filename); err != nil {
			return err
		}
	}
	return nil
}

func Mkdir(dir string, perm os.FileMode) error {
	if !IsDirExist(dir) {
		if err := os.MkdirAll(dir, perm); err != nil {
			return err
		}
	}
	return nil
}

// IsDir returns whether the filename is directory.
func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// IsExist returns whether the filename is exists.
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err) || err == nil
}

// IsNotExist returns whether the filename is exists.
func IsNotExist(filename string) bool {
	_, err := os.Stat(filename)
	return os.IsNotExist(err)
}

// IsDirExist reports whether dir exists and which is directory.
func IsDirExist(dir string) bool {
	fi, err := os.Stat(dir)
	return err == nil && fi.IsDir()
}

// IsGoFile returns whether the filename is exists.
func IsGoFile(filename string) bool {
	f, err := os.Stat(filename)
	return err == nil && filepath.Ext(f.Name()) == ".go"
}
