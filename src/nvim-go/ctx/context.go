// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctx

import (
	"go/build"
	"os"
	"path/filepath"
	"sync"

	"nvim-go/internal/pathutil"

	"github.com/neovim/go-client/nvim"
)

// Context represents a current nvim instances context.
type Context struct {
	// Errlist map the nvim quickfix errors.
	Errlist map[string][]*nvim.QuickfixError

	prevDir string // for cache
	m       sync.Mutex

	Buffer
	Build
}

// Buffer represents a buffer context.
type Buffer struct {
	// BufNr number of current buffer.
	BufNr int
	// WinID id of current window.
	WinID int
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

// SetContext sets the Tool, ProjectRoot, go/build.Default and $GOPATH to buildContext.
// This function initializes for functions that use go/build.Default.
func (ctx *Context) SetContext(dir string) {
	if dir != "" && ctx.prevDir != dir {
		ctx.m.Lock()
		defer ctx.m.Unlock()

		ctx.Build.Tool, ctx.Build.ProjectRoot, build.Default = buildContext(dir, build.Default)
		if ctx.Build.Tool == "gb" {
			build.Default.JoinPath = ctx.Build.GbJoinPath
		}
		ctx.prevDir = dir

		os.Setenv("GOPATH", build.Default.GOPATH)
	}
}
