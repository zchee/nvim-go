// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

func FindVCSRoot(root string) string {
	var vcsDirs = []string{".git", ".svn", ".hg", "_darcs"}
	if !IsDir(root) {
		root = filepath.Dir(root)
	}

	var found bool
	for {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() == false {
				return nil
			}
			for _, d := range vcsDirs {
				_, err := os.Stat(filepath.Join(path, d))
				if err == nil && strings.Contains(path, root) {
					found = true
					root = path
					break
				}
			}
			return nil
		})
		if found {
			break
		}
		root = filepath.Dir(root)
	}

	return filepath.Clean(root)
}
