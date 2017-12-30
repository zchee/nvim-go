package astmanip

import (
	"bytes"
	"reflect"
	"testing"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func TestCopyNode_Stmt(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata/copynodestmt/copynodestmt.go", nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	ast.Inspect(f, func(node ast.Node) bool {
		if _, ok := node.(ast.Stmt); ok {
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, node)
			deepCopied(t, buf.String(), node, CopyNode(node))
		}

		return true
	})
}

func TestCopyNode_Expr(t *testing.T) {
	exprs := []string{
		`x+y`,
		`x+1`,
		`foo(a, 1)`,
		`s[0]`,
		`s[1:]`,
		`s[1:3]`,
		`s[1:3:5]`,
		`(x)`,
		`foo.bar`,
		`*p`,
		`r.(*os.File)`,
		`-n`,
		`o.Meth(*p * 4, s[3:])`,
	}

	for _, expr := range exprs {
		e, err := parser.ParseExpr(expr)
		if err != nil {
			t.Fatal(err)
		}

		deepCopied(t, expr, e, CopyNode(e))
	}
}

func deepCopied(t *testing.T, name string, a, b ast.Node) {
	// ast.Node -> ast.BinaryExpr
	va := reflect.Indirect(reflect.ValueOf(a))
	vb := reflect.Indirect(reflect.ValueOf(b))

	typ := va.Type()

	if typ != vb.Type() {
		t.Errorf("type mismatch: %s %s", va, vb)
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fa := va.FieldByName(field.Name)
		fb := vb.FieldByName(field.Name)

		switch field.Type.Kind() {
		case reflect.Interface:
			// ast.Expr -> *ast.Ident
			fa = fa.Elem()
			fb = fb.Elem()

		case reflect.Ptr, reflect.Slice:
			// pass

		default:
			// non-pointer field such as token.Pos, token.Token and bool
			// t.Logf("DEBUG %s.%s %s: not a pointer type", typ, field.Name, field.Type)
			continue
		}

		if !fa.IsValid() && !fb.IsValid() {
			// nil interface
			// t.Logf("DEBUG %s.%s %s: not valid", typ, field.Name, field.Type)
			continue
		}

		if fa.Pointer() == 0 {
			// nil pointer
			continue
		}

		if fa.Pointer() == fb.Pointer() {
			t.Errorf("%s %q: fields equal: %s (%#v)", typ, name, field.Name, fa.Interface())
			return
		}

		// TODO: recurse
	}
}
