package testdata_nextsibling

import ( // <1> import
	"fmt"
)

var X string // <2> var

func F() { // <3> func
	foo() // <3.1>

	var b bool // <3.2>

	if true { // <3.3> if
		foo2() // <3.3.1>
		return // <3.3.2>
	}
}
