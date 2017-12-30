package astmanip

import (
	"fmt"
	"go/ast"
)

// CopyNode deep copies an ast.Node node and returns a new one.
func CopyNode(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.Ident:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Obj = nil
		return &copied

	case *ast.ArrayType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Len = copyExpr(node.Len)
		copied.Elt = copyExpr(node.Elt)
		return &copied

	case *ast.BadExpr:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.BasicLit:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.BinaryExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		copied.Y = copyExpr(node.Y)
		return &copied

	case *ast.CallExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Fun = copyExpr(node.Fun)
		copied.Args = copyExprSlice(node.Args)
		return &copied

	case *ast.ChanType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Value = copyExpr(node.Value)
		return &copied

	case *ast.CompositeLit:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Type = copyExpr(node.Type)
		copied.Elts = copyExprSlice(node.Elts)
		return &copied

	case *ast.Ellipsis:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Elt = copyExpr(node.Elt)
		return &copied

	case *ast.FuncLit:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Type = CopyNode(node.Type).(*ast.FuncType)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.FuncType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Params = copyFieldList(node.Params)
		copied.Results = copyFieldList(node.Results)
		return &copied

	case *ast.IndexExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		copied.Index = copyExpr(node.Index)
		return &copied

	case *ast.InterfaceType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Methods = copyFieldList(node.Methods)
		return &copied

	case *ast.KeyValueExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Key = copyExpr(node.Key)
		copied.Value = copyExpr(node.Value)
		return &copied

	case *ast.MapType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Key = copyExpr(node.Key)
		copied.Value = copyExpr(node.Value)
		return &copied

	case *ast.ParenExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		return &copied

	case *ast.SelectorExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		copied.Sel = copyIdent(node.Sel)
		return &copied

	case *ast.SliceExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		copied.Low = copyExpr(node.Low)
		copied.High = copyExpr(node.High)
		copied.Max = copyExpr(node.Max)
		return &copied

	case *ast.StarExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		return &copied

	case *ast.StructType:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Fields = copyFieldList(node.Fields)
		return &copied

	case *ast.TypeAssertExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		copied.Type = copyExpr(node.Type)
		return &copied

	case *ast.UnaryExpr:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		return &copied

	case *ast.AssignStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Lhs = copyExprSlice(node.Lhs)
		copied.Rhs = copyExprSlice(node.Rhs)
		return &copied

	case *ast.BadStmt:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.BlockStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.List = copyStmtSlice(node.List)
		return &copied

	case *ast.BranchStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Label = copyIdent(node.Label)
		return &copied

	case *ast.CaseClause:
		if node == nil {
			return nil
		}
		copied := *node
		copied.List = copyExprSlice(node.List)
		copied.Body = copyStmtSlice(node.Body)
		return &copied

	case *ast.CommClause:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Comm = copyStmt(node.Comm)
		copied.Body = copyStmtSlice(node.Body)
		return &copied

	case *ast.DeclStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Decl = CopyNode(node.Decl).(ast.Decl)

		return &copied

	case *ast.DeferStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Call = CopyNode(node.Call).(*ast.CallExpr)

		return &copied

	case *ast.EmptyStmt:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.ExprStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		return &copied

	case *ast.ForStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Init = copyStmt(node.Init)
		copied.Cond = copyExpr(node.Cond)
		copied.Post = copyStmt(node.Post)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.GoStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Call = CopyNode(node.Call).(*ast.CallExpr)

		return &copied

	case *ast.IfStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Init = copyStmt(node.Init)
		copied.Cond = copyExpr(node.Cond)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)
		copied.Else = copyStmt(node.Else)

		return &copied

	case *ast.IncDecStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.X = copyExpr(node.X)
		return &copied

	case *ast.LabeledStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Label = copyIdent(node.Label)
		copied.Stmt = copyStmt(node.Stmt)
		return &copied

	case *ast.RangeStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Key = copyExpr(node.Key)
		copied.Value = copyExpr(node.Value)
		copied.X = copyExpr(node.X)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.ReturnStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Results = copyExprSlice(node.Results)
		return &copied

	case *ast.SelectStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.SendStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Chan = copyExpr(node.Chan)
		copied.Value = copyExpr(node.Value)
		return &copied

	case *ast.SwitchStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Init = copyStmt(node.Init)
		copied.Tag = copyExpr(node.Tag)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.TypeSwitchStmt:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Init = copyStmt(node.Init)
		copied.Assign = copyStmt(node.Assign)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.ImportSpec:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Name = copyIdent(node.Name)
		copied.Path = CopyNode(node.Path).(*ast.BasicLit)
		copied.Comment = copyCommentGroup(node.Comment)

		return &copied

	case *ast.TypeSpec:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Name = copyIdent(node.Name)
		copied.Type = copyExpr(node.Type)
		copied.Comment = copyCommentGroup(node.Comment)
		return &copied

	case *ast.ValueSpec:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Names = copyIdentSlice(node.Names)
		copied.Type = copyExpr(node.Type)
		copied.Values = copyExprSlice(node.Values)
		copied.Comment = copyCommentGroup(node.Comment)
		return &copied

	case *ast.BadDecl:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.FuncDecl:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Recv = copyFieldList(node.Recv)
		copied.Name = copyIdent(node.Name)
		copied.Type = CopyNode(node.Type).(*ast.FuncType)
		copied.Body = CopyNode(node.Body).(*ast.BlockStmt)

		return &copied

	case *ast.GenDecl:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		if node.Specs == nil {
			copied.Specs = nil
		} else {
			copied.Specs = make([]ast.Spec, len(node.Specs))
			for i, x := range node.Specs {
				copied.Specs[i] =
					CopyNode(x).(ast.Spec)
			}
		}
		return &copied

	case *ast.Comment:
		if node == nil {
			return nil
		}
		copied := *node
		return &copied

	case *ast.CommentGroup:
		if node == nil {
			return nil
		}
		copied := *node
		if node.List == nil {
			copied.List = nil
		} else {
			copied.List = make([]*ast.Comment, len(node.List))
			for i, x := range node.List {
				copied.List[i] =
					CopyNode(x).(*ast.Comment)
			}
		}
		return &copied

	case *ast.Field:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Names = copyIdentSlice(node.Names)
		copied.Type = copyExpr(node.Type)
		copied.Tag = CopyNode(node.Tag).(*ast.BasicLit)
		copied.Comment = copyCommentGroup(node.Comment)

		return &copied

	case *ast.FieldList:
		if node == nil {
			return nil
		}
		copied := *node
		if node.List == nil {
			copied.List = nil
		} else {
			copied.List = make([]*ast.Field, len(node.List))
			for i, x := range node.List {
				copied.List[i] =
					CopyNode(x).(*ast.Field)
			}
		}
		return &copied

	case *ast.File:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Doc = copyCommentGroup(node.Doc)
		copied.Name = copyIdent(node.Name)
		if node.Decls == nil {
			copied.Decls = nil
		} else {
			copied.Decls = make([]ast.Decl, len(node.Decls))
			for i, x := range node.Decls {
				copied.Decls[i] =
					CopyNode(x).(ast.Decl)
			}
		}
		copied.Scope = nil
		if node.Imports == nil {
			copied.Imports = nil
		} else {
			copied.Imports = make([]*ast.ImportSpec, len(node.Imports))
			for i, x := range node.Imports {
				copied.Imports[i] =
					CopyNode(x).(*ast.ImportSpec)
			}
		}
		copied.Unresolved = copyIdentSlice(node.Unresolved)
		if node.Comments == nil {
			copied.Comments = nil
		} else {
			copied.Comments = make([]*ast.CommentGroup, len(node.Comments))
			for i, x := range node.Comments {
				copied.Comments[i] =
					CopyNode(x).(*ast.CommentGroup)
			}
		}
		return &copied

	case *ast.Package:
		if node == nil {
			return nil
		}
		copied := *node
		copied.Scope = nil
		copied.Files = copyFileMap(node.Files)
		return &copied

	default:
		fmt.Printf("CopyNode: unexpected node type %T\n", node)
		return node

	}
}

func copyExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}

	return CopyNode(expr).(ast.Expr)
}

func copyStmt(stmt ast.Stmt) ast.Stmt {
	if stmt == nil {
		return nil
	}

	return CopyNode(stmt).(ast.Stmt)
}

func copyDecl(decl ast.Decl) ast.Decl {
	if decl == nil {
		return nil
	}

	return CopyNode(decl).(ast.Decl)
}

func copyExprSlice(list []ast.Expr) []ast.Expr {
	if list == nil {
		return nil
	}

	copied := make([]ast.Expr, len(list))
	for i, expr := range list {
		copied[i] = copyExpr(expr)
	}
	return copied
}

func copyStmtSlice(list []ast.Stmt) []ast.Stmt {
	if list == nil {
		return nil
	}

	copied := make([]ast.Stmt, len(list))
	for i, stmt := range list {
		copied[i] = copyStmt(stmt)
	}
	return copied
}

func copyIdentSlice(list []*ast.Ident) []*ast.Ident {
	if list == nil {
		return nil
	}

	copied := make([]*ast.Ident, len(list))
	for i, ident := range list {
		copied[i] = copyExpr(ident).(*ast.Ident)
	}
	return copied
}

func copyCommentGroup(c *ast.CommentGroup) *ast.CommentGroup {
	if c == nil {
		return nil
	}

	return CopyNode(c).(*ast.CommentGroup)
}

func copySpecSlice(list []ast.Spec) []ast.Spec {
	if list == nil {
		return nil
	}

	copied := make([]ast.Spec, len(list))
	for i, spec := range list {
		copied[i] = CopyNode(spec).(ast.Spec)
	}
	return copied
}

func copyIdent(ident *ast.Ident) *ast.Ident {
	copied := CopyNode(ident)
	if copied == nil {
		return nil
	}

	return copied.(*ast.Ident)
}

func copyFieldList(fl *ast.FieldList) *ast.FieldList {
	if fl == nil {
		return nil
	}

	copied := *fl

	if fl.List != nil {
		copiedList := make([]*ast.Field, len(fl.List))
		for i, f := range fl.List {
			field := *f
			field.Names = make([]*ast.Ident, len(f.Names))
			for i, name := range f.Names {
				copiedName := *name
				field.Names[i] = &copiedName
			}
			field.Type = copyExpr(f.Type)
			copiedList[i] = &field
		}

		copied.List = copiedList
	}

	return &copied
}

func copyFileMap(m map[string]*ast.File) map[string]*ast.File {
	copied := map[string]*ast.File{}
	for k, v := range m {
		copied[k] = CopyNode(v).(*ast.File)
	}
	return copied
}
