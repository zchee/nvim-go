// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// FindVCSRoot finds the vcs root directory from path.
func FindVCSRoot(path string) string {
	var vcsDirs = []string{".git", ".svn", ".hg", "_darcs"}
	if !IsDir(path) {
		path = filepath.Dir(path)
	}

	var found bool
	for {
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() == false {
				return nil
			}
			for _, d := range vcsDirs {
				_, err := os.Stat(filepath.Join(p, d))
				if err == nil && strings.Contains(p, path) {
					found = true
					path = p
					break
				}
			}
			return nil
		})
		if err != nil {
			return filepath.Clean(path)
		}
		if found {
			break
		}
		path = filepath.Dir(path)
	}

	return filepath.Clean(path)
}
