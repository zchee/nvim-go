// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	testdataPath string
	testGbRoot   string
	testGoPath   string
)

func TestMain(m *testing.M) {
	var err error
	if testdataPath, err = filepath.Abs(filepath.Join("../testdata")); err != nil {
		log.Fatalf("pathutil_test: failed to get testdata directory: %v", err)
	}
	testGbRoot = filepath.Join(testdataPath, "gb")
	testGoPath = filepath.Join("testdata", "go")

	os.Exit(m.Run())
}

var buildContextMu sync.Mutex

func setBuildContext(t *testing.T, gopath string) func() {
	t.Helper()

	buildContextMu.Lock()
	defaulCtxt := build.Default
	build.Default.GOPATH = gopath

	return func() {
		build.Default = defaulCtxt
		buildContextMu.Unlock()
	}
}
