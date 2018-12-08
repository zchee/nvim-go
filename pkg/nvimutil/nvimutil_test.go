// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"os"
	"testing"

	"github.com/zchee/nvim-go/pkg/config"
)

func TestMain(m *testing.M) {
	_ = config.Process()

	os.Exit(m.Run())
	// goleak.VerifyTestMain(m)
}
