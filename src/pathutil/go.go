// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"go/build"
	"path/filepath"

	"github.com/pkg/errors"
)

// parsePackage search the parent directory of dir with the recursive loop.
func parsePackage(dir string) (*build.Package, error) {
	dir = filepath.Clean(dir)

	// for save the before(child) package information
	savePkg := new(build.Package)
	for {
		// Raise the error if dir is reaches root("/") or GOPATH or GOROOT
		if dir == "/" || dir == build.Default.GOPATH || dir == build.Default.GOROOT {
			return nil, errors.New("couldn't find the package")
		}

		// Get the current dir package information
		pkg, err := build.Default.ImportDir(dir, build.ImportMode(0))
		if err != nil {
			// Check the exists .go file in the dir
			if _, ok := err.(*build.NoGoError); ok {
				// Case of recursive loop of second and subsequent,
				// Use child(before) directory if NoGoError
				if savePkg.Dir != "" {
					return savePkg, nil
				} else {
					savePkg = pkg
					dir = filepath.Dir(dir)
					continue
				}
			} else if err != nil { // Failed ImportDir() on other errors
				return nil, errors.WithStack(err)
			}
		}

		// return the current directory if package is main
		if pkg.IsCommand() {
			return pkg, nil
		}

		// Return the child(before) package directory if different to current package name
		if beforePkgName := savePkg.Name; beforePkgName != "" && beforePkgName != pkg.Name {
			return savePkg, nil
		}

		// Save the current package and re-assign dir to parent dir for the next recursive loop
		savePkg = pkg
		dir = filepath.Dir(dir)
	}
}

// PackagePath returns the *full path* of package directory estimated
// from the dir directory structure.
// like:
//  return "/Users/zchee/go/src/github.com/pkg/errors", nil
func PackagePath(dir string) (string, error) {
	pkg, err := parsePackage(dir)
	if err != nil {
		return "", err
	}

	return pkg.Dir, nil
}

// PackageID returns the package ID(ImportPath) estimated from the dir
// directory structure.
// like:
//  return "github.com/pkg/errors", nil
func PackageID(dir string) (string, error) {
	pkg, err := parsePackage(dir)
	if err != nil {
		return "", err
	}

	return pkg.ImportPath, nil
}

// PackageRoot finds repository root of package from path.
// Return the '/root' if path is '/root/foo/bar'.
// If path package directory archtecture uses '/root/src/foo', returns the '/root/src'.
func PackageRoot(path string) (*build.Package, error) {
	if !IsDir(path) {
		path = filepath.Dir(path)
	}
	path = filepath.Clean(path)

	// for save the before(child) package information
	savePkg := new(build.Package)
	for {
		// Raise the error if dir is reaches root("/") or GOPATH or GOROOT
		if path == "/" || path == build.Default.GOPATH || path == build.Default.GOROOT {
			return nil, errors.New("couldn't find the package")
		}

		// Get the current dir package information
		pkg, err := build.Default.ImportDir(path, build.ImportMode(0))
		if err != nil {
			// Check the exists .go file in the dir
			if _, ok := err.(*build.NoGoError); ok {
				// Case of recursive loop of second and subsequent,
				// Use child(before) directory if NoGoError
				if savePkg.Dir != "" {
					return savePkg, nil
				} else {
					savePkg = pkg
					path = filepath.Dir(path)
					continue
				}
			} else if err != nil { // Failed ImportDir() on other errors
				return nil, errors.WithStack(err)
			}
		}

		// return the current directory if package is main
		if pkg.IsCommand() {
			return pkg, nil
		}

		// Return the child(before) package directory if different to current package name
		if beforePkgName := savePkg.Name; beforePkgName != "" && beforePkgName != pkg.Name {
			return savePkg, nil
		}

		// Save the current package and re-assign dir to parent dir for the next recursive loop
		savePkg = pkg
		path = filepath.Dir(path)
	}
}
