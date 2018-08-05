// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil_test

import (
	"log"
	"os"
	"path/filepath"
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
