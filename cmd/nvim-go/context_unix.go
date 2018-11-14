// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

var terminationSignals = []os.Signal{unix.SIGTERM, unix.SIGINT}
