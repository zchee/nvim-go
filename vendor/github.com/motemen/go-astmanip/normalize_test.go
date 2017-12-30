package astmanip

import (
	"bytes"
	"io/ioutil"
	"testing"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

func TestNormalizePos(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata/normalizepos/ugly.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	NormalizePos(f)
	source := formatNode(fset, f)
	b, err := ioutil.ReadFile("testdata/normalizepos/clean.go")
	if err != nil {
		t.Fatal(err)
	}

	if source != string(b) {
		t.Fatalf("source must be clean: \n%s---\n%s", source, string(b))
	}
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, node)
	return buf.String()
}
