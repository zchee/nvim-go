// Copyright 2017 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package home

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Dir detects and returns the user home directory.
func Dir() string {
	// At first, Check the $HOME environment variable
	if usrHome := os.Getenv("HOME"); usrHome != "" {
		return usrHome
	}

	// Fallback if not set $HOME
	// gets the canonical username
	cmdWhoami := exec.Command("whoami")
	usrName, err := cmdWhoami.Output()
	if err != nil {
		return ""
	}

	// gets the home directory path use 'eval echo ~$USER' magic
	stdout := new(bytes.Buffer)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("eval echo ~%s", usrName))
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}
