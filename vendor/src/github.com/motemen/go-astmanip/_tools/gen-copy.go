package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

// Generates type switch case clauses' bodies

func main() {
	overwrite := flag.Bool("w", false, "overwrite file")
	flag.Parse()

	conf := loader.Config{}
	conf.ParserMode = parser.ParseComments

	err := conf.CreateFromFilenames("", "copy.go")
	dieIf(err)

	prog, err := conf.Load()
	dieIf(err)

	pkg := prog.Created[0] // must be astutil
	file := pkg.Files[0]   // must be copy.go

	var copyNodeFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "CopyNode" {
				copyNodeFunc = funcDecl
				break
			}
		}
	}

	if copyNodeFunc == nil {
		log.Fatal("CopyNode() not found")
	}

	ast.Inspect(copyNodeFunc, func(node ast.Node) bool {
		clause, ok := node.(*ast.CaseClause)
		if !ok {
			return true
		}

		// the default case
		if clause.List == nil {
			return false
		}

		// assumption: len(clause.List) == 1
		caseType := pkg.TypeOf(clause.List[0])

		// assumption: case types are pointers to struct types
		baseStruct := caseType.(*types.Pointer).Elem().Underlying().(*types.Struct)

		typeCopier := map[types.Type]string{}
		var astNodeType *types.Interface

		{
			typeCopierMap := map[string]string{
				"ast.Node":               "CopyNode",
				"ast.Expr":               "copyExpr",
				"ast.Stmt":               "copyStmt",
				"*ast.CommentGroup":      "copyCommentGroup",
				"*ast.FieldList":         "copyFieldList",
				"*ast.Ident":             "copyIdent",
				"[]ast.Stmt":             "copyStmtSlice",
				"[]ast.Expr":             "copyExprSlice",
				"[]*ast.Ident":           "copyIdentSlice",
				"*ast.Object":            "nil",
				"*ast.Scope":             "nil",
				"map[string]*ast.File":   "copyFileMap",
				"map[string]*ast.Object": "",
			}

			for t, copier := range typeCopierMap {
				tv, err := types.Eval(t, pkg.Pkg, pkg.Scopes[file])
				dieIf(err)

				typeCopier[tv.Type] = copier

				if t == "ast.Node" {
					astNodeType = tv.Type.Underlying().(*types.Interface)
				}
			}
		}

		code := `package E
func x(node *struct{}) *struct{} {
	if node == nil {
		return nil
	}
	copied := *node
	// generated code goes here
	return &copied
}`

		f, err := parser.ParseFile(token.NewFileSet(), "", code, parser.Mode(0))
		dieIf(err)

		body := f.Decls[0].(*ast.FuncDecl).Body

		start := body.List[0 : len(body.List)-1]
		end := body.List[len(body.List)-1]

		clause.Body = start

		for i := 0; i < baseStruct.NumFields(); i++ {
			field := baseStruct.Field(i)
			if _, isBasic := field.Type().Underlying().(*types.Basic); isBasic {
				continue
			}

			var copier string = "TODO"
			var assertType ast.Expr
			for typ, c := range typeCopier {
				if types.Identical(field.Type(), typ) {
					copier = c
					break
				}
			}
			if copier == "TODO" {
				fieldType := field.Type()
				if types.ConvertibleTo(fieldType, astNodeType) {
					copier = "CopyNode"
					assertType, err = parser.ParseExpr(shortTypeString(fieldType))
					dieIf(err)
				} else if sliceType, ok := fieldType.(*types.Slice); ok {
					copySlice := genCopySlice(sliceType.Elem(), field.Name())
					clause.Body = append(clause.Body, copySlice...)
					continue
				} else {
					log.Println("no copier associated:", field.Name(), field.Type())
				}
			}

			// eg. CopyNode(node.X)
			var rhs ast.Expr

			if copier == "" {
				continue
			} else if copier == "nil" {
				rhs = ast.NewIdent("nil")
			} else {
				rhs = &ast.CallExpr{
					Fun: ast.NewIdent(copier),
					Args: []ast.Expr{
						&ast.SelectorExpr{
							ast.NewIdent("node"),
							ast.NewIdent(field.Name()),
						},
					},
				}
				if assertType != nil {
					// eg. CopyNode(node.X).(ast.Expr)
					rhs = &ast.TypeAssertExpr{
						X:    rhs,
						Type: assertType,
					}
				}
			}

			// eg. copied.X = CopyNode(node.X).(ast.Expr)
			assign := &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						ast.NewIdent("copied"),
						ast.NewIdent(field.Name()),
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{rhs},
			}

			clause.Body = append(clause.Body, assign)
		}

		clause.Body = append(clause.Body, end)

		return false
	})

	out := os.Stdout
	if *overwrite {
		var err error
		out, err = os.Create("copy.go")
		dieIf(err)
	}

	printer.Fprint(out, conf.Fset, file)
}

func shortTypeString(t types.Type) (s string) {
	if pt, ok := t.(*types.Pointer); ok {
		s = "*"
		t = pt.Elem()
	}
	obj := t.(*types.Named).Obj()
	s = s + obj.Pkg().Name() + "." + obj.Name()

	return
}

func genCopySlice(baseType types.Type, fieldName string) []ast.Stmt {
	code := `package E

func x(copied interface{}, node ast.Node) {
	if node.F == nil {
		copied.F = nil
	} else {
		copied.F = make([]T, len(node.F))
		for i, x := range node.F {
			copied.F[i] = CopyNode(x).(T)
		}
	}
}
`
	rewrites := map[string]string{
		"T": shortTypeString(baseType),
		"F": fieldName,
	}

	f, err := parser.ParseFile(token.NewFileSet(), "", code, parser.Mode(0))
	dieIf(err)

	body := f.Decls[0].(*ast.FuncDecl).Body
	ast.Inspect(body, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			if r, ok := rewrites[ident.Name]; ok {
				ident.Name = r
				ident.NamePos = token.NoPos
			}
			return false
		}

		return true
	})

	return body.List
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
