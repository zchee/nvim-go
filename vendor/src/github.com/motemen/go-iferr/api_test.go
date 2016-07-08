package iferr

import (
	"bytes"
	"os"
	"testing"

	"go/ast"
	"go/printer"

	_ "go/types"

	"golang.org/x/tools/go/loader"
)

func TestRewriteFile(t *testing.T) {
	conf := loader.Config{}

	source := `package P

import "log"

func e1() error {
	return nil
}

func b() error {
	err := e1()

	return nil
}

func b2() (int, error) {
	err := e1()

	return 100, nil
}

func c() {
	err := e1()
}
	`

	conf.AllowErrors = true

	f, err := conf.ParseFile("p.go", source)
	if err != nil {
		t.Fatal(err)
	}

	conf.CreateFromFiles("p", f)
	prog, err := conf.Load()
	if err != nil {
		t.Fatal(err)
	}

	RewriteFile(conf.Fset, f, prog.Package("p").Info)
	printer.Fprint(os.Stderr, conf.Fset, f)

	_ = err
}

func TestMakeZeroValue(t *testing.T) {
	for _, c := range []struct {
		defs string
		typ  string
	}{
		{"", "int8"},
		{"", "bool"},
		{"", "string"},
		{"", "complex128"},
		{"", "float64"},
		{"", "byte"},
		{"", "rune"},
		{"", "map[string]bool"},
		{"", "interface{}"},
		{"", "[]byte"},
		{"", "[8]byte"},
		{"", "struct{s string; b bool}"},
		{"", "error"},
		{"", "func() error"},
		{"", "chan bool"},
		{"", "<-chan bool"},
		{"type A struct{}", "A"},
		{"type A struct{}", "*A"},
		{"type A [8]byte", "A"},
		{`import "io"`, "io.Reader"},
		{`import "os"`, "os.FileMode"},
	} {
		source := "package P; "
		if c.defs != "" {
			source += c.defs + "; "
		}
		source += "var x " + c.typ

		conf := loader.Config{}
		f, err := conf.ParseFile("testdata.go", source)
		if err != nil {
			t.Fatal(err)
		}

		conf.CreateFromFiles("P", f)
		prog, err := conf.Load()
		if err != nil {
			t.Fatal(err)
		}

		pkg := prog.Package("P")
		file := pkg.Files[0]

		var (
			typeExpr  = file.Decls[len(file.Decls)-1].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type
			typesType = pkg.Pkg.Scope().Lookup("x").Type()
		)

		zeroValue := makeZeroValue(typeExpr, typesType)

		buf := bytes.NewBufferString(source + " = ")
		printer.Fprint(buf, conf.Fset, zeroValue)
		t.Log(buf.String())

		g, err := conf.ParseFile("testdata_2.go", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		conf.CreateFromFiles("P_2", g)
		_, err = conf.Load()
		if err != nil {
			t.Fatal(err)
		}
	}
}
