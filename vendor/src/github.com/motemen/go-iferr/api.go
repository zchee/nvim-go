// TODO(motemen): Don't guess the variable name to "err"
// TODO(motemen): Make error handling code custamizable

package iferr

import (
	"fmt"
	"log"

	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/motemen/go-astmanip"
)

var (
	panicCode    = "panic(err.Error())"
	logFatalCode = "log.Fatal(err)"
	tFatalCode   = "t.Fatal(err)"
)

var errorType types.Type

func init() {
	errorType = types.Universe.Lookup("error").Type()
	log.SetFlags(log.Lshortfile)
}

// errorAssign is an assign statement which involves an error-typed variable.
type errorAssign struct {
	outerFunc *ast.FuncDecl
	stmt      *ast.AssignStmt
	ident     *ast.Ident
}

func RewriteFile(fset *token.FileSet, f *ast.File, info types.Info) {
	errAssigns := []errorAssign{}

	ast.Inspect(f, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
				t := info.TypeOf(ident)
				if t == nil {
					log.Printf("%s: could not detect type of %s", fset.Position(ident.Pos()), ident.Name)
					continue
				}
				if types.Identical(t, errorType) {
					var funcDecl *ast.FuncDecl
					path, _ := astutil.PathEnclosingInterval(f, assign.Pos(), assign.End())
					for _, p := range path {
						var ok bool
						funcDecl, ok = p.(*ast.FuncDecl)
						if ok {
							break
						}
					}
					if funcDecl != nil {
						errAssigns = append(errAssigns, errorAssign{
							outerFunc: funcDecl,
							stmt:      assign,
							ident:     ident,
						})
					}
					break
				}
			}
		}

		return false
	})

	for _, assign := range errAssigns {
		assignLine := fset.Position(assign.stmt.Pos()).Line
		next := astmanip.NextSibling(f, assign.stmt)
		if next == nil || fset.Position(next.Pos()).Line-assignLine > 1 {
			catch := makeErrorCatchStatement(
				assign.ident, makeErrorHandleStatement(assign, info),
			)
			astmanip.InsertStmtAfter(assign.outerFunc.Body, catch, assign.stmt)
		}
	}
}

func makeErrorHandleStatement(assign errorAssign, info types.Info) ast.Stmt {
	if funcResults := assign.outerFunc.Type.Results; funcResults != nil {
		errorPosInReturnTypes := -1
		for i, rt := range funcResults.List {
			if types.Identical(info.TypeOf(rt.Type), errorType) {
				errorPosInReturnTypes = i
				break
			}
		}
		if errorPosInReturnTypes != -1 {
			returnValues := make([]ast.Expr, len(funcResults.List))
			for i, rt := range funcResults.List {
				if i == errorPosInReturnTypes {
					// return ..., err, ...
					returnValues[i] = ast.NewIdent(assign.ident.Name)
				} else {
					// return ..., zv, ...
					zv := makeZeroValue(rt.Type, info.TypeOf(rt.Type))
					returnValues[i] = zv
				}
			}
			return &ast.ReturnStmt{Results: returnValues}
		}
	}

	var code string

	funcScope := info.Scopes[assign.outerFunc.Type]
	if tVar, ok := funcScope.Lookup("t").(*types.Var); ok {
		if tVarType, ok := tVar.Type().(*types.Pointer); ok {
			if tVarType, ok := tVarType.Elem().(*types.Named); ok {
				tVarTypeObj := tVarType.Obj()
				if tVarTypeObj.Pkg().Path() == "testing" && tVarTypeObj.Name() == "T" {
					code = tFatalCode
				}
			}
		}
	}
	if code == "" {
		_, logObj := info.Scopes[assign.outerFunc.Type].LookupParent("log", token.NoPos)
		if logPkg, ok := logObj.(*types.PkgName); ok && logPkg.Imported().Path() == "log" {
			code = logFatalCode
		}
	}
	if code == "" {
		code = panicCode
	}

	expr, err := parser.ParseExpr(code)
	if err != nil {
		panic(fmt.Sprintf("must not fail: %s while parsing %q", err, code))
	}

	return &ast.ExprStmt{X: expr}
}

var ifTemplate = `package _; func _() { if err != nil {} }`

func makeErrorCatchStatement(errName *ast.Ident, stmt ast.Stmt) *ast.IfStmt {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "var ifTemplate", ifTemplate, 0)
	if err != nil {
		panic(fmt.Sprintf("must not fail: %s", err))
	}

	// must not fail
	ifStmt := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.IfStmt)
	ifStmt.Body.List = []ast.Stmt{stmt}

	astmanip.NormalizePos(ifStmt)

	return ifStmt
}

func makeZeroValue(e ast.Expr, t types.Type) ast.Expr {
	switch t := t.(type) {
	case *types.Basic:
		switch {
		case t.Info()&types.IsNumeric != 0:
			return &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			}

		case t.Info()&types.IsString != 0:
			return &ast.BasicLit{
				Kind:  token.STRING,
				Value: `""`,
			}

		case t.Info()&types.IsBoolean != 0:
			return ast.NewIdent("false")
		}

		panic(fmt.Sprintf("makeZeroValue: unexpected basic type: %v", t))

	case *types.Tuple:
		panic("makeZeroValue: unexpected *types.Tuple")

	case *types.Named:
		return makeZeroValue(e, t.Underlying())

	case *types.Array:
		return &ast.CompositeLit{Type: e}
	case *types.Struct:
		return &ast.CompositeLit{Type: e}

	case *types.Map:
		return ast.NewIdent("nil")
	case *types.Signature:
		return ast.NewIdent("nil")
	case *types.Interface:
		return ast.NewIdent("nil")
	case *types.Pointer:
		return ast.NewIdent("nil")
	case *types.Slice:
		return ast.NewIdent("nil")
	case *types.Chan:
		return ast.NewIdent("nil")
	}

	panic(fmt.Sprintf("makeZeroValue: unexpected type: %v", t))
}
