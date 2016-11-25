// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type FindMode int

const (
	ModeExcludeVendor FindMode = 1 << iota
)

// FindAllPackage returns a list of all packages in all of the GOPATH trees
// in the given build context. If prefix is non-empty, only packages
// whose import paths begin with prefix are returned.
func FindAllPackage(root string, buildContext build.Context, ignores []string, mode FindMode) ([]*build.Package, error) {
	var (
		pkgs []*build.Package
		done = make(map[string]bool)
	)

	filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}

		// avoid .foo, _foo, and testdata directory trees.
		_, elem := filepath.Split(path)
		if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" || (mode&ModeExcludeVendor != 0 && elem == "vendor") || matchIgnore(elem, ignores) {
			return filepath.SkipDir
		}

		name := filepath.ToSlash(path[len(root):])
		if done[name] {
			return nil
		}
		done[name] = true

		pkg, err := buildContext.ImportDir(path, build.IgnoreVendor)
		if err != nil && strings.Contains(err.Error(), "no buildable Go source files") {
			return nil
		}
		pkgs = append(pkgs, pkg)
		return nil
	})
	return pkgs, nil
}

func matchIgnore(elem string, ignores []string) bool {
	for _, e := range ignores {
		if elem == e {
			return true
		}
	}
	return false
}
