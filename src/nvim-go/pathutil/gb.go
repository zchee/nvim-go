// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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
	if IsNotExist(vendor) {
		return dir, false
	}
	return root, true
}

// FindGbProjectRoot works upwards from path seaching for the src/ directory
// which identifies the project root.
// Code taken directly from constabulary/gb.
//  github.com/constabulary/gb/cmd/path.go
func FindGbProjectRoot(path string) (string, error) {
	if path == "" {
		return "", errors.New("project root is blank")
	}
	start := path
	for path != filepath.Dir(path) {
		root := filepath.Join(path, "src")
		if IsNotExist(root) {
			path = filepath.Dir(path)
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf(`could not find project root in "%s" or its parents`, start)
}

// GbProjectName return the gb project name.
func GbProjectName(projectRoot string) string {
	return filepath.Base(projectRoot)
}

func GbPackages(root string) ([]string, error) {
	dir := filepath.Join(root, "src")
	paths, err := ioutil.ReadDir(dir)
	if err != nil {
		errors.Wrapf(err, "could not read %s dir", dir)
	}
	pkgs := make([]string, 0, len(paths))
	for _, path := range paths {
		if path.IsDir() && !strings.HasPrefix(path.Name(), "_") {
			pkgs = append(pkgs, path.Name())
		}
	}

	return pkgs, nil
}
