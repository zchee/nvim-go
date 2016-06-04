// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sync"

	"github.com/juju/errors"
)

const pkgContext = "context"

// Build specifies the supporting context for a build and embedded
// build.Context type struct.
type Build struct {
	Tool         string
	GbProjectDir string
	build.Context
}

// GoPath return the new GOPATH estimated from the path p directory structure.
func (ctxt *Build) buildContext(p string) (string, string) {
	tool := "go"

	// Get original $GOPATH path.
	goPath := os.Getenv("GOPATH")

	// Check the path p are Gb directory structure.
	// If ok, append gb root and vendor path to the goPath lists.
	if gbpath, ok := ctxt.isGb(p); ok {
		ctxt.GbProjectDir = gbpath
		goPath = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
		tool = "gb"
	}

	return goPath, tool
}

// contextMu Mutex lock for SetContext.
var contextMu sync.Mutex

// SetContext sets the go/build Default.GOPATH and $GOPATH to GoPath(p)
// under a mutex.
// The returned function restores Default.GOPATH to its original value and
// unlocks the mutex.
//
// This function intended to be used to the go/build Default.
func (ctxt *Build) SetContext(p string) func() {
	contextMu.Lock()
	original := build.Default.GOPATH

	ctxt.Context = build.Default

	build.Default.GOPATH, ctxt.Tool = ctxt.buildContext(p)
	ctxt.Context.GOPATH = build.Default.GOPATH
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default.GOPATH = original
		ctxt.Context.GOPATH = build.Default.GOPATH
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}

func (ctxt *Build) PackageDir(dir string) (string, error) {
	dir = filepath.Clean(dir)

	savePkg := new(build.Package)
	for {
		// Get the current files package information
		pkg, err := ctxt.ImportDir(dir, build.IgnoreVendor)
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

		// Save the current package name
		savePkg = pkg
		dir = filepath.Dir(dir)
	}

	err := errors.Errorf("cannot find the package path from %s", dir)
	return "", err
}

// isGb check the current buffer directory whether gb directory structure.
// Return the gb project root path and boolean, and sets the context.GbProjectDir.
func (ctxt *Build) isGb(dir string) (string, bool) {
	root, err := FindGbProjectRoot(dir)
	if err != nil {
		return "", false
	}

	// FindGbProjectRoot Gets the GOPATH root if go directory structure.
	// Recheck use vendor directory.
	vendor := filepath.Join(root, "vendor")
	if _, err := os.Stat(vendor); err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
	}
	return root, true
}

// FindGbProjectRoot works upwards from path seaching for the
// src/ directory which identifies the project root.
// Code taken directly from gb.
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
