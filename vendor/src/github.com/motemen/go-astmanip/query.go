package astmanip

import (
	"go/ast"
)

// NextSibling eturns the node right after ref, inside the root tree.
// If ref is the last child of its parent, then returns nil.
func NextSibling(root, ref ast.Node) (result ast.Node) {
	type state uint
	const (
		stateInitial state = iota
		stateCaptureNext
		stateDone
	)

	var st state
	ast.Inspect(root, func(node ast.Node) bool {
		switch st {
		case stateInitial:
			if node == ref {
				st++
				return false
			}
			return true

		case stateCaptureNext:
			result = node
			st++
			return false

		case stateDone:
			return false
		}

		panic("unreachable")
	})

	return
}
