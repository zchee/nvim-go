package context

import (
	"os"
	"path/filepath"
)

var (
	vcsDirs     = []string{".git", ".svn", ".hg"}
	vcsDirFound bool
)

// FindVcsDir find package root path from arg path
func FindVcsDir(basedir string) string {
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
