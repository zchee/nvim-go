// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import "os"

func IsDebug() bool {
	return os.Getenv("NVIM_GO_DEBUG") != ""
}
