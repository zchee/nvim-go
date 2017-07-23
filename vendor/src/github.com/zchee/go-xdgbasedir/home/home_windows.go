// Copyright 2017 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package home

import (
	"os"
	"path/filepath"
)

// Dir detects and returns the user home directory.
func Dir() string {
	// At first, Check the $HOME environment variable
	usrHome := os.Getenv("HOME")
	if usrHome != "" {
		return filepath.FromSlash(usrHome)
	}

	// TODO(zchee): In Windows OS, which of $HOME and these checks has priority?
	// Respect the USERPROFILE environment variable because Go stdlib uses it for default GOPATH in the "go/build" package.
	if usrHome = os.Getenv("USERPROFILE"); usrHome == "" {
		usrHome = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}

	return filepath.FromSlash(usrHome)
}
