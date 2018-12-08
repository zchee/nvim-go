// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strutil

import (
	"reflect"
	"runtime"
	"unsafe"
)

// noescape hides a pointer from escape analysis.
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

// UnsafeSlice returns a byte array that points to the given string without a heap allocation.
// The string must be preserved until the  byte arrayis disposed.
func UnsafeSlice(s string) (p []byte) {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	sh.Data = (*(*reflect.StringHeader)(noescape(unsafe.Pointer(&s)))).Data
	sh.Len = len(s)
	sh.Cap = len(s)

	runtime.KeepAlive(&s)
	return p
}

// UnsafeString returns a string that points to the given byte array without a heap allocation.
// The byte array must be preserved until the string is disposed.
func UnsafeString(p []byte) (s string) {
	if len(p) == 0 {
		return
	}

	(*reflect.StringHeader)(unsafe.Pointer(&s)).Data = uintptr(noescape(unsafe.Pointer(&p[0])))
	(*reflect.StringHeader)(unsafe.Pointer(&s)).Len = len(p)

	runtime.KeepAlive(&p)
	return
}
