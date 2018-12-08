// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package version

// tag is version of nvim-go.
//
// This is used for GCP profiler Sentry error reporting and so on. tag is overridden using
// `-X main.tag` during release builds.
var Tag string

// gitCommit is commit hash of nvim-go.
//
// This is used for GCP profiler Sentry error reporting and so on. gitCommit is overridden using
// `-X main.gitCommit` during release builds.
var GitCommit string

var Version = Tag + "@" + GitCommit

// var Version string
