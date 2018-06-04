// Copyright 2018 The go-xdgbasedir Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package xdgbasedir implements a freedesktop.org XDG Base Directory Specification.
//  https://specifications.freedesktop.org/basedir-spec/latest/
//
// The XDG Base Directory Specification is based on the following concepts:
//
// - There is a single base directory relative to which user-specific data files should be written. This directory is defined by the environment variable $XDG_DATA_HOME.
//
// - There is a single base directory relative to which user-specific configuration files should be written. This directory is defined by the environment variable $XDG_CONFIG_HOME.
//
// - There is a set of preference ordered base directories relative to which data files should be searched. This set of directories is defined by the environment variable $XDG_DATA_DIRS.
//
// - There is a set of preference ordered base directories relative to which configuration files should be searched. This set of directories is defined by the environment variable $XDG_CONFIG_DIRS.
//
// - There is a single base directory relative to which user-specific non-essential (cached) data should be written. This directory is defined by the environment variable $XDG_CACHE_HOME.
package xdgbasedir // import "github.com/zchee/go-xdgbasedir"
