// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"sync"
)

const pkgContext = "context"

type Context struct {
	context.Context

	Build BuildContext
}

// Build specifies the supporting context for a build and embedded build.Context type struct.
type BuildContext struct {
	Tool         string
	GbProjectDir string
}

// GoPath return the new GOPATH estimated from the path p directory structure.
func (ctxt *BuildContext) buildContext(p string) build.Context {
	buildDefault := build.Default

	ctxt.Tool = "go"
	// Get original $GOPATH path.
	goPath := os.Getenv("GOPATH")

	// Check the path p are Gb directory structure.
	// If ok, append gb root and vendor path to the goPath lists.
	if gbpath, ok := ctxt.isGb(filepath.Clean(p)); ok {
		ctxt.GbProjectDir = gbpath
		ctxt.Tool = "gb"
		buildDefault.JoinPath = ctxt.GbJoinPath

		goPath = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
	}

	buildDefault.GOPATH = goPath

	return buildDefault
}

// contextMu Mutex lock for SetContext.
var contextMu sync.Mutex

// SetContext sets the go/build Default.GOPATH and $GOPATH to GoPath(p)
// under a mutex.
// The returned function restores Default.GOPATH to its original value and
// unlocks the mutex.
//
// This function intended to be used to the go/build Default.
func (ctxt *BuildContext) SetContext(p string) func() {
	contextMu.Lock()
	original := build.Default

	build.Default = ctxt.buildContext(p)
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default = original
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}
