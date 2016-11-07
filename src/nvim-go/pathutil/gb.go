// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsGb check the current buffer directory whether gb directory structure.
// Return the gb project root path and boolean, and sets the context.GbProjectDir.
func IsGb(dir string) (string, bool) {
	root, err := FindGbProjectRoot(dir)
	if err != nil {
		return "", false
	}

	// Check root directory whether "vendor", and overwrite root path to parent of vendor directory.
	if filepath.Base(root) == "vendor" {
		root = filepath.Dir(root)
	}
	// FindGbProjectRoot gets the GOPATH root if go directory structure.
	// Recheck use vendor directory.
	vendor := filepath.Join(root, "vendor")
	if _, err := os.Stat(vendor); err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
	}
	return root, true
}

// FindGbProjectRoot works upwards from path seaching for the src/ directory
// which identifies the project root.
// Code taken directly from constabulary/gb.
//  github.com/constabulary/gb/cmd/path.go
func FindGbProjectRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("project root is blank")
	}
	start := path
	for path != filepath.Dir(path) {
		root := filepath.Join(path, "src")
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				path = filepath.Dir(path)
				continue
			}
			return "", err
		}
		return path, nil
	}
	return "", fmt.Errorf(`could not find project root in "%s" or its parents`, start)
}

// GbProjectName return the gb project name.
func GbProjectName(projectRoot string) string {
	return filepath.Base(projectRoot)
}
