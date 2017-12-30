package astmanip

import (
	"fmt"
	"strings"
	"testing"

	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
)

var commentedNodeCache = map[string]ast.Node{}

func commentedNode(fset *token.FileSet, f *ast.File, comment string) ast.Node {
	if comment == "" {
		return nil
	}

	if n, ok := commentedNodeCache[comment]; ok {
		return n
	}

	for _, c := range f.Comments {
		if strings.TrimSpace(c.Text()) == comment {
			pos := c.Pos()
			// move position to line start
			pos = token.Pos(int(pos) - fset.Position(pos).Column + 1)
			path, _ := astutil.PathEnclosingInterval(f, pos, c.Pos())

			var node ast.Node
			for n := 0; n < len(path); n++ {
				if n == len(path)-1 {
					panic(fmt.Sprintf("cannot find commented node: %q", comment))
				}

				if path[n].Pos() == path[n+1].Pos() && path[n].End() == path[n+1].End() {
					continue
				}

				node = path[n]

				break
			}

			commentedNodeCache[comment] = node
			return node
		}
	}

	panic(fmt.Sprintf("cannot find commented node: %q", comment))
}

func TestNextSibling(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata/nextsibling/nextsibling.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	mustBeNextSibling := func(name1, name2 string) {
		node1 := commentedNode(fset, f, name1)
		node2 := commentedNode(fset, f, name2)
		next := NextSibling(f, node1)
		if next != node2 {
			t.Fatalf("node %q %#v must be next sibling to %q %#v, got %#v", name2, node2, name1, node1, next)
		}
	}

	mustBeNextSibling("<1> import", "<2> var")
	mustBeNextSibling("<2> var", "<3> func")
	mustBeNextSibling("<3> func", "")
	mustBeNextSibling("<3.1>", "<3.2>")
	mustBeNextSibling("<3.2>", "<3.3> if")
	mustBeNextSibling("<3.3> if", "")
	mustBeNextSibling("<3.3.1>", "<3.3.2>")
	mustBeNextSibling("<3.3.2>", "")
}
