// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
)

var (
	vcsDirs     = []string{".git", ".svn", ".hg"}
	foundVCSDir bool
)

// FindVCSRoot find package root path from arg path
func FindVCSRoot(basedir string) string {
	foundVCSDir = false

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

func findvcsDirWalkFunc(path string, fileInfo os.FileInfo, err error) error {
	if err != nil || fileInfo == nil || fileInfo.IsDir() == false {
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
