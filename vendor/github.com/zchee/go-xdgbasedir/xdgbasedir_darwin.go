// Copyright 2017 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package xdgbasedir

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/zchee/go-xdgbasedir/home"
)

var (
	defaultDataHome   string
	defaultConfigHome string
	defaultDataDirs   string
	defaultConfigDirs string
	defaultCacheHome  string
	defaultRuntimeDir string
)

func initDir() {
	initOnce.Do(func() {
		switch Mode {
		case Unix:
			defaultDataHome = filepath.Join(home.Dir(), ".local", "share")
			defaultConfigHome = filepath.Join(home.Dir(), ".config")
			defaultDataDirs = filepath.Join("/usr", "local", "share") + string(filepath.ListSeparator) + filepath.Join("/usr", "share")
			defaultConfigDirs = filepath.Join("/etc", "xdg")
			defaultCacheHome = filepath.Join(home.Dir(), ".cache")
			defaultRuntimeDir = filepath.Join("/run", "user", strconv.Itoa(os.Getuid()))
		case Native:
			// ref: https://developer.apple.com/library/content/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/MacOSXDirectories/MacOSXDirectories.html
			defaultDataHome = filepath.Join(home.Dir(), "Library", "Application Support")
			defaultConfigHome = filepath.Join(home.Dir(), "Library", "Preferences")
			defaultDataDirs = defaultDataHome
			defaultConfigDirs = defaultConfigHome
			defaultCacheHome = filepath.Join(home.Dir(), "Library", "Caches")
			defaultRuntimeDir = defaultDataHome
		}
	})
}

func dataHome() string {
	initDir()
	return defaultDataHome
}

func configHome() string {
	initDir()
	return defaultConfigHome
}

func dataDirs() string {
	initDir()
	return defaultDataDirs
}

func configDirs() string {
	initDir()
	return defaultConfigDirs
}

func cacheHome() string {
	initDir()
	return defaultCacheHome
}

func runtimeDir() string {
	initDir()
	return defaultRuntimeDir
}
