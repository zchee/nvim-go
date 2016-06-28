// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"path/filepath"

	"github.com/juju/errors"
)

// PackagePath returns the import path of the current file of the package.
func (ctxt *BuildContext) PackagePath(dir string) (string, error) {
	dir = filepath.Clean(dir)

	savePkg := new(build.Package)
	for {
		// Get the current files package information
		pkg, err := build.Default.ImportDir(dir, build.IgnoreVendor)
		// noGoError := &build.NoGoError{Dir: dir}
		if _, ok := err.(*build.NoGoError); ok {
			// if err == noGoError {
			return savePkg.ImportPath, nil
		} else if err != nil {
			return "", errors.Annotate(err, pkgContext)
		}

		if pkg.IsCommand() {
			return pkg.ImportPath, nil
		} else if savePkg.Name != "" && pkg.Name != savePkg.Name {
			return savePkg.ImportPath, nil
		}

		if dir == "/" {
			return "", errors.Errorf("cannot find the package path from %s", dir)
		}

		// Save the current package name
		savePkg = pkg
		dir = filepath.Dir(dir)
	}
}
