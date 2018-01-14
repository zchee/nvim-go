// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"go/build"
	"sync"
	"testing"
)

var buildContextMu sync.Mutex

func SetBuildContext(t *testing.T, gopath string) func() {
	t.Helper()

	buildContextMu.Lock()
	defaulCtxt := build.Default
	build.Default.GOPATH = gopath

	return func() {
		build.Default = defaulCtxt
		buildContextMu.Unlock()
	}
}
