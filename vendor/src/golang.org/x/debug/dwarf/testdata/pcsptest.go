// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func main() {
	test(1, 2, 3)
}

// This is the function we examine. After the preamble its stack should be
// pulled down 1*addrSize for the return PC plus 3*8 for the three
// arguments. That will be (1+3)*8=32 on 64-bit machines.
func test(a, b, c int64) int64 {
	// Put in enough code that it's not inlined.
	for a = 0; a < 100; a++ {
		b += c
	}
	afterTest(a, b, c)
	return b
}

// This function follows test in the binary. We use it to force arguments
// onto the stack and as a delimiter in the text we scan in the test.
func afterTest(a, b, c int64) {
	fmt.Println(a, b, c)
}
