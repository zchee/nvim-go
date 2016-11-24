// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neovim/go-client/nvim"
)

type xdgEnv struct {
	env  string
	path string
}

// TestNvim sets XDG tmpdir and return the *nvim.Nvim for unit testing.
func TestNvim(t *testing.T, file ...string) *nvim.Nvim {
	tmpDir := filepath.Join(os.TempDir(), "nvim-go-test")
	defer os.RemoveAll(tmpDir)

	if err := setXDGEnv(tmpDir); err != nil {
		t.Fatalf("couldn't setup XDG directories: %v", err)
	}

	// Disable debug log output
	os.Setenv("NVIM_GO_DEBUG", "")

	// -u: Use <init.vim> instead of the default
	// -n: No swap file, use memory only
	args := []string{"-u", "NONE", "-n"}
	if file != nil {
		// -p: Open N tab pages (default: one for each file)
		args = append(args, "-p")

		for _, f := range file {
			args = append(args, f)
		}
	}
	n, err := nvim.NewEmbedded(&nvim.EmbedOptions{
		Args: args,
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- n.Serve()
		err := n.Close()
		serveErr := <-done
		if err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatal(serveErr)
		}
	}()

	return n
}

// setXDGEnv create and sets the XDG_* directories to the environment based by
// freedesktop.org XDG Base Directory Specification.
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func setXDGEnv(dir string) error {
	home := filepath.Join(dir, "home", "testnvim")
	if err := os.MkdirAll(home, 0755); err != nil {
		return err
	}

	xdgEnvDir := []xdgEnv{
		{"XDG_CONFIG_HOME", filepath.Join(home, ".config")},
		{"XDG_CACHE_HOME", filepath.Join(home, ".cache")},
		{"XDG_DATA_HOME", filepath.Join(home, ".local", "share")},
		{"XDG_RUNTIME_DIR", filepath.Join(dir, "run", "user", "1000")},
		{"XDG_DATA_DIRS", filepath.Join(dir, "usr", "local", "share")},
		{"XDG_CONFIG_DIRS", filepath.Join(dir, "etc", "xdg")},
	}
	for _, p := range xdgEnvDir {
		os.Setenv(p.env, p.path)
		if err := os.MkdirAll(p.path, 0755); err != nil {
			return err
		}
	}

	return nil
}
