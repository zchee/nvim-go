// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"os"
	"path/filepath"
	"strings"

	"nvim-go/context"
)

func RelPath(f, cwd string) string {
	if filepath.HasPrefix(f, cwd) {
		return strings.Replace(f, cwd+string(filepath.Separator), "", 1)
	}
	rel, _ := filepath.Rel(cwd, f)
	return rel
}

func PackagePath(p string) string {
	goPath := os.Getenv("GOPATH")
	p = context.FindVcsRoot(p)

	return strings.Replace(p, goPath+string(filepath.Separator), "", 1)
}

func GbProjectName(p string, ctxt *context.Build) string {
	pkgPath := strings.Replace(p, filepath.Join(ctxt.ProjectDir, "src")+string(filepath.Separator), "", 1)
	return strings.Split(pkgPath, string(filepath.Separator))[0]
}
