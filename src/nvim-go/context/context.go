// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"sync"
)

// A Context specifies the supporting context for a build and embedded
// build.Context type struct.
type Build struct {
	Tool       string
	ProjectDir string
	build.Context
}

// GoPath return the new GOPATH estimated from the path p directory structure.
func (ctxt *Build) buildContext(p string) (string, string) {
	tool := "go"

	// Get original $GOPATH path.
	goPath := os.Getenv("GOPATH")

	// Check the path p are Gb directory structure.
	// If ok, append gb root and vendor path to the goPath lists.
	if gbpath, ok := ctxt.isGb(filepath.Clean(p)); ok {
		goPath = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
		tool = "gb"
	}

	return goPath, tool
}

// isGb return the gb package root path if p is gb project directory structure.
func (ctxt *Build) isGb(p string) (string, bool) {
	ctxt.Context = build.Default

	// First check
	manifest := filepath.Join(p, "vendor/manifest")
	if _, err := os.Stat(manifest); err == nil {
		return p, true
	}

	var pkgRoot string
	for {
		pkg, _ := ctxt.ImportDir(p, build.IgnoreVendor)
		rootDir := FindVcsRoot(p)

		if pkg.Name == "main" || rootDir == pkg.Dir {
			pkgRoot = pkg.Dir
			break
		}
		p = filepath.Dir(p)
		continue
	}

	projRoot := pkgRoot
	src := "src"
	if projRoot != FindVcsRoot(p) {
		// gb project directory is `../../pkgRoot`
		projRoot, src = filepath.Split(filepath.Dir(pkgRoot))
		if src != "src" {
			return "", false
		}
	}

	manifest = filepath.Join(filepath.Clean(projRoot), "vendor/manifest")
	_, err := os.Stat(manifest)
	if err != nil {
		return "", false
	}
	ctxt.ProjectDir = filepath.Clean(projRoot)

	return ctxt.ProjectDir, (err == nil && src == "src")
}

// contextMu Mutex lock for SetContext.
var contextMu sync.Mutex

// SetContext sets the go/build Default.GOPATH and $GOPATH to GoPath(p)
// under a mutex.
// The returned function restores Default.GOPATH to its original value and
// unlocks the mutex.
//
// This function intended to be used to the go/build Default.
func (c *Build) SetContext(p string) func() {
	contextMu.Lock()
	original := build.Default.GOPATH

	build.Default.GOPATH, c.Tool = c.buildContext(p)
	c.Context.GOPATH = build.Default.GOPATH
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default.GOPATH = original
		c.Context.GOPATH = build.Default.GOPATH
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}
