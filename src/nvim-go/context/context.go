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

	"github.com/neovim-go/vim"
	"golang.org/x/net/context"
)

const pkgContext = "context"

// Context represents a embeded context package and build context.
type Context struct {
	context.Context
	Build

	Errlist map[string][]*vim.QuickfixError
}

// Build represents a compile tool information.
type Build struct {
	// Tool name of project compile tool
	Tool string
	// ProjectRoot package import path in the case of go project, GB_PROJECT_DIR in the case of
	// gb project.
	ProjectRoot string
}

// NewContext return the Context type with initialize Context.Errlist.
func NewContext() *Context {
	return &Context{
		Errlist: make(map[string][]*vim.QuickfixError),
	}
}

// buildContext return the new build context estimated from the path p directory structure.
func (ctxt *Context) buildContext(p string) build.Context {
	buildContext := build.Default

	ctxt.Build.Tool = "go"
	ctxt.Build.ProjectRoot, _ = pathutil.PackagePath(p)
	// Get original $GOPATH path.
	buildContext.GOPATH = os.Getenv("GOPATH")

	// Check the path p are Gb directory structure.
	// If ok, append gb root and vendor path to the goPath lists.
	if gbpath, ok := pathutil.IsGb(filepath.Clean(p)); ok {
		ctxt.Build.Tool = "gb"
		ctxt.Build.ProjectRoot = gbpath
		buildContext.JoinPath = ctxt.GbJoinPath
		buildContext.GOPATH = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
	}

	return buildContext
}

// contextMu Mutex lock for SetContext.
var contextMu sync.Mutex

// SetContext sets the go/build Default.GOPATH and $GOPATH to GoPath(p)
// under a mutex.
// The returned function restores Default.GOPATH to its original value and
// unlocks the mutex.
//
// This function intended to be used to the go/build Default.
func (ctxt *Context) SetContext(p string) func() {
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
