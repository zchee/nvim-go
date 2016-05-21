// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"os"
	"path/filepath"
)

var (
	vcsDirs     = []string{".git", ".svn", ".hg"}
	vcsDirFound bool
)

// FindVcsRoot find package root path from arg path
func FindVcsRoot(basedir string) string {
	vcsDirFound = false
	filepath.Walk(basedir, findvcsDirWalkFunc)

	for {
		if !vcsDirFound {
			basedir = filepath.Dir(basedir)
			filepath.Walk(basedir, findvcsDirWalkFunc)
		} else {
			break
		}
	}

	return basedir
}

func findvcsDirWalkFunc(path string, fileInfo os.FileInfo, err error) error {
	if err != nil || fileInfo == nil || fileInfo.IsDir() == false {
		return nil
	}

	for _, d := range vcsDirs {
		_, err := os.Stat(filepath.Join(path, d))
		if err == nil {
			vcsDirFound = true
			break
		}
	}

	return nil
}
