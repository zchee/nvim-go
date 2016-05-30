// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"sync"

	"nvim-go/pathutil"
)

// Build specifies the supporting context for a build and embedded
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
	if gbpath, ok := ctxt.isGb(p); ok {
		goPath = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
		tool = "gb"
	}

	return goPath, tool
}

// isGb check the current buffer directory whether gb directory structure.
// Return the gb project root path and boolean, and sets the context.GbProjectDir.
func (ctxt *Build) isGb(dir string) (string, bool) {
	root, err := pathutil.FindGbProjectRoot(dir)
	if err != nil {
		return "", false
	}

	// pathutil.FindGbProjectRoot Gets the GOPATH root if go directory structure.
	// Recheck use vendor directory.
	vendor := filepath.Join(pathutil.FindVcsRoot(dir), "vendor")
	if _, err := os.Stat(vendor); err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
	}
	ctxt.ProjectDir = root
	return root, true
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
