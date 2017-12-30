// TODO(motemen): tests

// Package astmanip provides miscellaneous functions for go/ast manipulations.
package astmanip

import (
	"fmt"
	"go/ast"
)

// InsertStmtAfter inserts a statement stmt after ref, inside the root tree.
// Specifically it updates []ast.Stmt inside BlockStmt, CaseClause, or CommClause.
func InsertStmtAfter(root ast.Node, stmt, ref ast.Stmt) {
	var done bool
	ast.Inspect(root, func(n ast.Node) bool {
		if done {
			return false
		}

		parentStmt, ok := n.(ast.Stmt)
		if !ok {
			return true
		}

		var found bool
		ast.Inspect(parentStmt, func(n ast.Node) bool {
			if n == parentStmt {
				return true
			}
			if n == ref {
				found = true
			}
			return false
		})
		if found {
			insertStmtAfter(parentStmt, stmt, ref)
			done = true
		}

		return !done
	})

	if !done {
		panic("cannot find parent")
	}
}

func insertStmtAfter(parent ast.Node, node, ref ast.Stmt) {
	switch p := parent.(type) {
	case *ast.BlockStmt:
		p.List = insertStmtIntoListAfter(p.List, node, ref)
	case *ast.CaseClause:
		p.Body = insertStmtIntoListAfter(p.Body, node, ref)
	case *ast.CommClause:
		p.Body = insertStmtIntoListAfter(p.Body, node, ref)
	default:
		panic(fmt.Sprintf("unexpected parent node: %T", parent))
	}
}

func insertStmtIntoListAfter(list []ast.Stmt, stmt, ref ast.Stmt) []ast.Stmt {
	for i, s := range list {
		if s == ref {
			return append(list[0:i+1], append([]ast.Stmt{stmt}, list[i+1:]...)...)
		}
	}

	panic("could not find ref stmt")
}
