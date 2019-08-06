// Copyright 2019 The color Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package color

import (
	"reflect"
	"unsafe"
)

// unsafeToSlice returns a byte array that points to the given string without a heap allocation.
// The string must be preserved until the byte array is disposed.
func unsafeToSlice(s string) (p []byte) {
	if s == "" {
		return
	}

	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	p = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}))

	return
}
