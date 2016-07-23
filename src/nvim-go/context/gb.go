// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
)

// GbJoinPath joins the sequence of path fragments into a single path for build.Default.JoinPath.
func (ctxt *Build) GbJoinPath(elem ...string) string {
	res := filepath.Join(elem...)

	if gbrel, err := filepath.Rel(ctxt.ProjectRoot, res); err == nil {
		gbrel = filepath.ToSlash(gbrel)
		gbrel, _ = match(gbrel, "vendor/")
		if gbrel, ok := match(gbrel, fmt.Sprintf("pkg/%s_%s", build.Default.GOOS, build.Default.GOARCH)); ok {
			gbrel, hasSuffix := match(gbrel, "_")

			if hasSuffix {
				gbrel = "-" + gbrel
			}
			gbrel = fmt.Sprintf("pkg/%s-%s/", build.Default.GOOS, build.Default.GOARCH) + gbrel
			gbrel = filepath.FromSlash(gbrel)
			res = filepath.Join(ctxt.ProjectRoot, gbrel)
		}
	}

	return res
}

func match(s, prefix string) (string, bool) {
	rest := strings.TrimPrefix(s, prefix)
	return rest, len(rest) < len(s)
}
