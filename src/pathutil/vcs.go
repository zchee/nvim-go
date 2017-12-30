// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
)

var vcsDirs = []string{".git", ".svn", ".hg"}

// FindVCSRoot find package root path from arg path
func FindVCSRoot(basedir string) string {
	var foundVCSDir bool

	findvcsDirWalkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() == false {
			return nil
		}

		for _, d := range vcsDirs {
			_, err := os.Stat(filepath.Join(path, d))
			if err == nil {
				foundVCSDir = true
				break
			}
		}

		return nil
	}

	for {
		filepath.Walk(basedir, findvcsDirWalkFunc)
		if !foundVCSDir {
			basedir = filepath.Dir(basedir)
			continue
		}
		break
	}

	return filepath.Clean(basedir)
}
