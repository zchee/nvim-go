// Copyright 2017 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xdgbasedir

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type mode int

const (
	// Unix unix mode directory structure.
	Unix mode = iota
	// Native native mode directory structure.
	Native
)

// Mode mode of directory structure. This config only available darwin.
//
// If it is set to `Unix`, it refers to the same path as linux. If it is set to `Native`, it refers to the Apple FileSystemProgrammingGuide path.
// By default, `Unix`.
var Mode = Unix

// initOnce for run initDir once on darwin.
var initOnce sync.Once

// DataHome return the XDG_DATA_HOME based directory path.
//
// $XDG_DATA_HOME defines the base directory relative to which user specific data files should be stored.
// If $XDG_DATA_HOME is either not set or empty, a default equal to $HOME/.local/share should be used.
func DataHome() string {
	if env := os.Getenv("XDG_DATA_HOME"); env != "" {
		return expandUser(env)
	}
	return dataHome()
}

// ConfigHome return the XDG_CONFIG_HOME based directory path.
//
// $XDG_CONFIG_HOME defines the base directory relative to which user specific configuration files should be stored.
// If $XDG_CONFIG_HOME is either not set or empty, a default equal to $HOME/.config should be used.
func ConfigHome() string {
	if env := os.Getenv("XDG_CONFIG_HOME"); env != "" {
		return expandUser(env)
	}
	return configHome()
}

// DataDirs return the XDG_DATA_DIRS based directory path.
//
// $XDG_DATA_DIRS defines the preference-ordered set of base directories to search for data files in addition
// to the $XDG_DATA_HOME base directory. The directories in $XDG_DATA_DIRS should be seperated with a colon ':'.
// If $XDG_DATA_DIRS is either not set or empty, a value equal to /usr/local/share/:/usr/share/ should be used.
func DataDirs() string {
	if env := os.Getenv("XDG_DATA_DIRS"); env != "" {
		return expandUser(env)
	}
	return dataDirs()
}

// ConfigDirs return the XDG_CONFIG_DIRS based directory path.
//
// $XDG_CONFIG_DIRS defines the preference-ordered set of base directories to search for configuration files in addition
// to the $XDG_CONFIG_HOME base directory. The directories in $XDG_CONFIG_DIRS should be seperated with a colon ':'.
// If $XDG_CONFIG_DIRS is either not set or empty, a value equal to /etc/xdg should be used.
func ConfigDirs() string {
	if env := os.Getenv("XDG_CONFIG_DIRS"); env != "" {
		return expandUser(env)
	}
	return configDirs()
}

// CacheHome return the XDG_CACHE_HOME based directory path.
//
// $XDG_CACHE_HOME defines the base directory relative to which user specific non-essential data files should be stored.
// If $XDG_CACHE_HOME is either not set or empty, a default equal to $HOME/.cache should be used.
func CacheHome() string {
	if env := os.Getenv("XDG_CACHE_HOME"); env != "" {
		return expandUser(env)
	}
	return cacheHome()
}

// RuntimeDir return the XDG_RUNTIME_DIR based directory path.
//
// $XDG_RUNTIME_DIR defines the base directory relative to which user-specific non-essential runtime files and
// other file objects (such as sockets, named pipes, ...) should be stored. The directory MUST be owned by the user,
// and he MUST be the only one having read and write access to it. Its Unix access mode MUST be 0700.
//
// TODO(zchee): XDG_RUNTIME_DIR seems to change depending on the each distro or init system such as systemd.
// Also In macOS, normal user haven't permission for write to this directory.
// xref:
//	http://serverfault.com/questions/388840/good-default-for-xdg-runtime-dir/727994#727994
func RuntimeDir() string {
	if env := os.Getenv("XDG_RUNTIME_DIR"); env != "" {
		return expandUser(env)
	}
	return runtimeDir()
}

// expandUser expands shell's user home directory tilde expansion from s.
func expandUser(s string) string {
	if len(s) < 2 || s[0] != '~' || !os.IsPathSeparator(s[1]) {
		return s
	}

	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	home := os.Getenv(env)
	if home == "" {
		return s
	}

	if runtime.GOOS == "windows" {
		s = filepath.ToSlash(filepath.Join(home, s[2:]))
	} else {
		s = filepath.Join(home, s[2:])
	}
	return os.Expand(s, func(env string) string {
		if env == "HOME" {
			return home
		}
		return os.Getenv(env)
	})
}
