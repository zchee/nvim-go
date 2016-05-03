// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// A Context specifies the supporting context for a build and embedded
// build.Context type struct.
type Build struct {
	Tool string
	build.Context
}

// GoPath return the new GOPATH estimated from the path p directory structure.
func (ctxt *Build) buildContext(p string) (string, string) {
	tool := "go"

	// Get original $GOPATH path.
	goPath := os.Getenv("GOPATH")

	// Get runtime $GOROOT path and join to goPath.
	r := runtime.GOROOT()
	if r != "" {
		goPath = goPath + string(filepath.ListSeparator) + r
	}

	// Cleanup directory path.
	p = filepath.Clean(p)

	// Check the path p are Gb directory structure.
	// If yes, append gb root and vendor path to the goPath lists.
	if gbpath, yes := ctxt.isGb(p); yes {
		goPath = gbpath + string(filepath.ListSeparator) +
			filepath.Join(gbpath, "vendor") + string(filepath.ListSeparator) +
			goPath
		tool = "gb"
	}

	return goPath, tool
}

// isGb return the gb package root path if p is gb project directory structure.
func (ctxt *Build) isGb(p string) (string, bool) {
	ctxt.Context = build.Default

	var pkgRoot string
	for {
		pkg, err := ctxt.ImportDir(p, build.IgnoreVendor)
		if err != nil {
			return "", err != nil
			break
		}

		if pkg.Name == "main" {
			pkgRoot = pkg.Dir
			break
		}
		p = filepath.Dir(p)
		continue
	}

	// gb project directory is `../../pkgRoot`
	projRoot, src := filepath.Split(filepath.Dir(pkgRoot))

	_, err := os.Stat(filepath.Join(filepath.Clean(projRoot), "vendor/manifest"))

	return filepath.Clean(projRoot), (err != nil || src == "src")
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
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default.GOPATH = original
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}
