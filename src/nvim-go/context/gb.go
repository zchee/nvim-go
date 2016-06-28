// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// isGb check the current buffer directory whether gb directory structure.
// Return the gb project root path and boolean, and sets the context.GbProjectDir.
func (ctxt *BuildContext) isGb(dir string) (string, bool) {
	root, err := FindGbProjectRoot(dir)
	if err != nil {
		return "", false
	}

	// Check current directory whether "vendor", and overwrite root path
	if filepath.Base(root) == "vendor" {
		root = filepath.Dir(root)
	}
	// FindGbProjectRoot Gets the GOPATH root if go directory structure.
	// Recheck use vendor directory.
	vendor := filepath.Join(root, "vendor")
	if _, err := os.Stat(vendor); err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
	}
	return root, true
}

// FindGbProjectRoot works upwards from path seaching for the src/ directory
// which identifies the project root.
// Code taken directly from constabulary/gb.
//  github.com/constabulary/gb/cmd/path.go
func FindGbProjectRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("project root is blank")
	}
	start := path
	for path != filepath.Dir(path) {
		root := filepath.Join(path, "src")
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				path = filepath.Dir(path)
				continue
			}
			return "", err
		}
		return path, nil
	}
	return "", fmt.Errorf(`could not find project root in "%s" or its parents`, start)
}

func (ctxt *Build) GbJoinPath(elem ...string) string {
	res := filepath.Join(elem...)

	if gbrel, err := filepath.Rel(ctxt.GbProjectDir, res); err == nil {
		gbrel = filepath.ToSlash(gbrel)
		gbrel, _ = match(gbrel, "vendor/")
		if gbrel, ok := match(gbrel, fmt.Sprintf("pkg/%s_%s", ctxt.BuildContext.GOOS, ctxt.BuildContext.GOARCH)); ok {
			gbrel, hasSuffix := match(gbrel, "_")

			if hasSuffix {
				gbrel = "-" + gbrel
			}
			gbrel = fmt.Sprintf("pkg/%s-%s/", ctxt.BuildContext.GOOS, ctxt.BuildContext.GOARCH) + gbrel
			gbrel = filepath.FromSlash(gbrel)
			res = filepath.Join(ctxt.GbProjectDir, gbrel)
		}
	}

	return res
}

func match(s, prefix string) (string, bool) {
	rest := strings.TrimPrefix(s, prefix)
	return rest, len(rest) < len(s)
}
