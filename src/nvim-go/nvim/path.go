// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"path/filepath"
	"strings"
)

func ToRelPath(f, cwd string) string {
	if filepath.HasPrefix(f, cwd) {
		return strings.Replace(f, cwd+string(filepath.Separator), "", 1)
	}
	rel, _ := filepath.Rel(cwd, f)
	return rel
}
