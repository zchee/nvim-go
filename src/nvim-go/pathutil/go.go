// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"go/build"
	"path/filepath"

	"github.com/pkg/errors"
)

// PackagePath returns the package full directory path estimated from the path p directory structure.
// return "/Users/zchee/go/src/github.com/pkg/errors", nil
// TODO(zchee): duplicate function behavior of PackageID.
func PackagePath(dir string) (string, error) {
	dir = filepath.Clean(dir)

	savePkg := new(build.Package)
	for {
		// Get the current files package information
		pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
		// noGoError := &build.NoGoError{Dir: dir}
		if _, ok := err.(*build.NoGoError); ok {
			// if err == noGoError {
			return savePkg.Dir, nil
		} else if err != nil {
			return "", errors.Wrap(err, pkgPathutil)
		}

		if savePkg.Name != "" && pkg.Name != savePkg.Name {
			return savePkg.Dir, nil
		} else if pkg.IsCommand() {
			return pkg.Dir, nil
		}

		if dir == "/" {
			return "", errors.Errorf("cannot find the package path from %s", dir)
		}

		// Save the current package name
		savePkg = pkg
		dir = filepath.Dir(dir)
	}
}

// PackageID returns the package ID estimated from the path p directory structure.
//  return "github.com/pkg/errors", nil
// TODO(zchee): duplicate function behavior of PackagePath.
func PackageID(dir string) (string, error) {
	savePkg := new(build.Package)
	for {
		// Get the current files package information
		pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
		// noGoError := &build.NoGoError{Dir: dir}
		if _, ok := err.(*build.NoGoError); ok {
			// if err == noGoError {
			return savePkg.ImportPath, nil
		} else if err != nil {
			return "", errors.Wrap(err, pkgPathutil)
		}

		if savePkg.Name != "" && pkg.Name != savePkg.Name {
			return savePkg.ImportPath, nil
		} else if pkg.IsCommand() {
			return pkg.ImportPath, nil
		}

		if dir == "/" {
			return "", errors.Errorf("cannot find the package path from %s", dir)
		}

		// Save the current package name
		savePkg = pkg
		dir = filepath.Dir(dir)
	}
}
