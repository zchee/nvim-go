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

// A Context specifies the supporting context for a build and
// embedded build.Context type struct.
type Build struct {
	Tool string
	build.Context
}

// GoPath return the new GOPATH estimated from the path p directory structure.
func (ctxt *Build) GoPath(p string) string {
	ctxt.Tool = "go"

	// Get original $GOPATH path.
	goPath := os.Getenv("GOPATH")

	// Get runtime $GOROOT path and join to goPath.
	r := runtime.GOROOT()
	if r != "" {
		goPath = goPath + string(filepath.ListSeparator) + r
	}

	// Cleanup directory path.
	p = filepath.Clean(p)

	// path p is Gb directory structure? if yes, append gb root and vendor path
	// to the goPath lists.
	if gbpath, yes := ctxt.isGb(p); yes {
		goPath = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor") + string(filepath.ListSeparator) + goPath
		ctxt.Tool = "gb"
	}

	return goPath
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

	projRoot, src := filepath.Split(filepath.Dir(pkgRoot))

	_, err := os.Stat(filepath.Join(filepath.Clean(projRoot), "vendor/manifest"))

	return projRoot, (err != nil || src == "src")
}

// contextMu Mutex lock for SetContext
var contextMu sync.Mutex

func SetContext(p string) func() {
	contextMu.Lock()
	original := build.Default.GOPATH

	var c Build
	build.Default.GOPATH = c.GoPath(p)
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default.GOPATH = original
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}
