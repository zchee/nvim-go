// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neovim/go-client/nvim"
)

var (
	testCwd, _   = os.Getwd()
	testdataPath = filepath.Join(testCwd, "../testdata")
	testGoPath   = filepath.Join(testdataPath, "go")
	testGbPath   = filepath.Join(testdataPath, "gb")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
	broken      = filepath.Join(testGoPath, "src", "broken")
	brokenMain  = filepath.Join(broken, "broken.go")
	gsftp       = filepath.Join(testGbPath, "gsftp", "src", "cmd", "gsftp")
	gsftpRoot   = filepath.Join(testCwd, "testdata", "gb", "gsftp")
	gsftpMain   = filepath.Join(gsftpRoot, "src", "cmd", "gsftp", "main.go")
)

func benchVim(b *testing.B, file string) *nvim.Nvim {
	tmpdir := filepath.Join(os.TempDir(), "nvim-go-test")
	setXDGEnv(tmpdir)
	defer os.RemoveAll(tmpdir)

	os.Setenv("NVIM_GO_DEBUG", "")

	// -u: Use <init.vim> instead of the default
	// -n: No swap file, use memory only
	nvimArgs := []string{"-u", "NONE", "-n"}
	if file != "" {
		nvimArgs = append(nvimArgs, file)
	}
	v, err := nvim.NewEmbedded(&nvim.EmbedOptions{
		Args: nvimArgs,
		Logf: b.Logf,
	})
	if err != nil {
		b.Fatal(err)
	}

	go v.Serve()
	return v
}

func setXDGEnv(tmpdir string) {
	xdgDir := filepath.Join(tmpdir, "xdg")
	os.MkdirAll(xdgDir, 0)

	os.Setenv("XDG_RUNTIME_DIR", xdgDir)
	os.Setenv("XDG_DATA_HOME", xdgDir)
	os.Setenv("XDG_CONFIG_HOME", xdgDir)
	os.Setenv("XDG_DATA_DIRS", xdgDir)
	os.Setenv("XDG_CONFIG_DIRS", xdgDir)
	os.Setenv("XDG_CACHE_HOME", xdgDir)
	os.Setenv("XDG_LOG_HOME", xdgDir)
}
