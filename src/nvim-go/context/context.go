// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"sync"

	"nvim-go/pathutil"

	"github.com/neovim/go-client/nvim"
)

// Context represents a embeded context package and build context.
type Context struct {
	Build

	Errlist map[string][]*nvim.QuickfixError
}

// Build represents a build tool information.
type Build struct {
	// Tool name of build tool
	Tool string
	// ProjectRoot package directory full path in the case of go project,
	// GB_PROJECT_DIR in the case of gb project.
	ProjectRoot string
}

// NewContext return the Context type with initialize Context.Errlist.
func NewContext() *Context {
	return &Context{
		Errlist: make(map[string][]*nvim.QuickfixError),
	}
}

// buildContext return the new build context estimated from the path p directory structure.
func buildContext(dir string, defaultContext build.Context) (string, string, build.Context) {
	// copy context
	buildContext := defaultContext

	// Default is go context
	tool := "go"
	// Assign package directory full path from dir
	projectRoot, _ := pathutil.PackagePath(dir)

	// Check whether the dir is Gb directory structure.
	// If ok, append gb root and vendor path to the goPath lists.
	if gbpath, ok := pathutil.IsGb(filepath.Clean(dir)); ok {
		tool = "gb"
		projectRoot = gbpath
		buildContext.GOPATH = gbpath + string(filepath.ListSeparator) + filepath.Join(gbpath, "vendor")
	}

	return tool, projectRoot, buildContext
}

// contextMu Mutex lock for SetContext.
var contextMu sync.Mutex

// SetContext sets the go/build Default.GOPATH and $GOPATH to GoPath(p)
// under a mutex.
// The returned function restores Default.GOPATH to its original value and
// unlocks the mutex.
//
// This function intended to be used to the go/build Default.
func (ctx *Context) SetContext(dir string) func() {
	contextMu.Lock()
	original := build.Default

	ctx.Build.Tool, ctx.Build.ProjectRoot, build.Default = buildContext(dir, original)
	if ctx.Build.Tool == "gb" {
		build.Default.JoinPath = ctx.Build.GbJoinPath
	}
	os.Setenv("GOPATH", build.Default.GOPATH)

	return func() {
		build.Default = original
		os.Setenv("GOPATH", build.Default.GOPATH)
		contextMu.Unlock()
	}
}
