package astmanip

import (
	"go/ast"
	"go/token"
)

func normalize(p *token.Pos) {
	if *p != token.NoPos {
		*p = 1
	}
}

// NormalizePos resets all position information of node and its descendants.
func NormalizePos(node ast.Node) {
	if node.Pos() == token.NoPos && node.End() == token.NoPos {
		return
	}

	switch n := node.(type) {
	case *ast.ArrayType:
		normalize(&n.Lbrack)
		if n.Len != nil {
			NormalizePos(n.Len)
		}
		if n.Elt != nil {
			NormalizePos(n.Elt)
		}

	case *ast.AssignStmt:
		for _, x := range n.Lhs {
			NormalizePos(x)
		}
		normalize(&n.TokPos)
		for _, x := range n.Rhs {
			NormalizePos(x)
		}

	case *ast.BadDecl:
		normalize(&n.From)
		normalize(&n.To)

	case *ast.BadExpr:
		normalize(&n.From)
		normalize(&n.To)

	case *ast.BadStmt:
		normalize(&n.From)
		normalize(&n.To)

	case *ast.BasicLit:
		normalize(&n.ValuePos)

	case *ast.BinaryExpr:
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.OpPos)
		if n.Y != nil {
			NormalizePos(n.Y)
		}

	case *ast.BlockStmt:
		normalize(&n.Lbrace)
		for _, x := range n.List {
			NormalizePos(x)
		}
		normalize(&n.Rbrace)

	case *ast.BranchStmt:
		normalize(&n.TokPos)
		if n.Label != nil {
			NormalizePos(n.Label)
		}

	case *ast.CallExpr:
		if n.Fun != nil {
			NormalizePos(n.Fun)
		}
		normalize(&n.Lparen)
		for _, x := range n.Args {
			NormalizePos(x)
		}
		normalize(&n.Ellipsis)
		normalize(&n.Rparen)

	case *ast.CaseClause:
		normalize(&n.Case)
		for _, x := range n.List {
			NormalizePos(x)
		}
		normalize(&n.Colon)
		for _, x := range n.Body {
			NormalizePos(x)
		}

	case *ast.ChanType:
		normalize(&n.Begin)
		normalize(&n.Arrow)
		if n.Value != nil {
			NormalizePos(n.Value)
		}

	case *ast.CommClause:
		normalize(&n.Case)
		if n.Comm != nil {
			NormalizePos(n.Comm)
		}
		normalize(&n.Colon)
		for _, x := range n.Body {
			NormalizePos(x)
		}

	case *ast.Comment:
		normalize(&n.Slash)

	case *ast.CommentGroup:
		for _, x := range n.List {
			NormalizePos(x)
		}

	case *ast.CompositeLit:
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		normalize(&n.Lbrace)
		for _, x := range n.Elts {
			NormalizePos(x)
		}
		normalize(&n.Rbrace)

	case *ast.DeclStmt:
		if n.Decl != nil {
			NormalizePos(n.Decl)
		}

	case *ast.DeferStmt:
		normalize(&n.Defer)
		if n.Call != nil {
			NormalizePos(n.Call)
		}

	case *ast.Ellipsis:
		normalize(&n.Ellipsis)
		if n.Elt != nil {
			NormalizePos(n.Elt)
		}

	case *ast.EmptyStmt:
		normalize(&n.Semicolon)

	case *ast.ExprStmt:
		if n.X != nil {
			NormalizePos(n.X)
		}

	case *ast.Field:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		for _, x := range n.Names {
			NormalizePos(x)
		}
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		if n.Tag != nil {
			NormalizePos(n.Tag)
		}
		if n.Comment != nil {
			NormalizePos(n.Comment)
		}

	case *ast.FieldList:
		normalize(&n.Opening)
		for _, x := range n.List {
			NormalizePos(x)
		}
		normalize(&n.Closing)

	case *ast.File:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		normalize(&n.Package)
		if n.Name != nil {
			NormalizePos(n.Name)
		}
		for _, x := range n.Decls {
			NormalizePos(x)
		}
		for _, x := range n.Imports {
			NormalizePos(x)
		}
		for _, x := range n.Unresolved {
			NormalizePos(x)
		}
		for _, x := range n.Comments {
			NormalizePos(x)
		}

	case *ast.ForStmt:
		normalize(&n.For)
		if n.Init != nil {
			NormalizePos(n.Init)
		}
		if n.Cond != nil {
			NormalizePos(n.Cond)
		}
		if n.Post != nil {
			NormalizePos(n.Post)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		if n.Recv != nil {
			NormalizePos(n.Recv)
		}
		if n.Name != nil {
			NormalizePos(n.Name)
		}
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.FuncLit:
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.FuncType:
		normalize(&n.Func)
		if n.Params != nil {
			NormalizePos(n.Params)
		}
		if n.Results != nil {
			NormalizePos(n.Results)
		}

	case *ast.GenDecl:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		normalize(&n.TokPos)
		normalize(&n.Lparen)
		for _, x := range n.Specs {
			NormalizePos(x)
		}
		normalize(&n.Rparen)

	case *ast.GoStmt:
		normalize(&n.Go)
		if n.Call != nil {
			NormalizePos(n.Call)
		}

	case *ast.Ident:
		normalize(&n.NamePos)

	case *ast.IfStmt:
		normalize(&n.If)
		if n.Init != nil {
			NormalizePos(n.Init)
		}
		if n.Cond != nil {
			NormalizePos(n.Cond)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}
		if n.Else != nil {
			NormalizePos(n.Else)
		}

	case *ast.ImportSpec:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		if n.Name != nil {
			NormalizePos(n.Name)
		}
		if n.Path != nil {
			NormalizePos(n.Path)
		}
		if n.Comment != nil {
			NormalizePos(n.Comment)
		}
		normalize(&n.EndPos)

	case *ast.IncDecStmt:
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.TokPos)

	case *ast.IndexExpr:
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.Lbrack)
		if n.Index != nil {
			NormalizePos(n.Index)
		}
		normalize(&n.Rbrack)

	case *ast.InterfaceType:
		normalize(&n.Interface)
		if n.Methods != nil {
			NormalizePos(n.Methods)
		}

	case *ast.KeyValueExpr:
		if n.Key != nil {
			NormalizePos(n.Key)
		}
		normalize(&n.Colon)
		if n.Value != nil {
			NormalizePos(n.Value)
		}

	case *ast.LabeledStmt:
		if n.Label != nil {
			NormalizePos(n.Label)
		}
		normalize(&n.Colon)
		if n.Stmt != nil {
			NormalizePos(n.Stmt)
		}

	case *ast.MapType:
		normalize(&n.Map)
		if n.Key != nil {
			NormalizePos(n.Key)
		}
		if n.Value != nil {
			NormalizePos(n.Value)
		}

	case *ast.Package:

	case *ast.ParenExpr:
		normalize(&n.Lparen)
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.Rparen)

	case *ast.RangeStmt:
		normalize(&n.For)
		if n.Key != nil {
			NormalizePos(n.Key)
		}
		if n.Value != nil {
			NormalizePos(n.Value)
		}
		normalize(&n.TokPos)
		if n.X != nil {
			NormalizePos(n.X)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.ReturnStmt:
		normalize(&n.Return)
		for _, x := range n.Results {
			NormalizePos(x)
		}

	case *ast.SelectStmt:
		normalize(&n.Select)
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.SelectorExpr:
		if n.X != nil {
			NormalizePos(n.X)
		}
		if n.Sel != nil {
			NormalizePos(n.Sel)
		}

	case *ast.SendStmt:
		if n.Chan != nil {
			NormalizePos(n.Chan)
		}
		normalize(&n.Arrow)
		if n.Value != nil {
			NormalizePos(n.Value)
		}

	case *ast.SliceExpr:
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.Lbrack)
		if n.Low != nil {
			NormalizePos(n.Low)
		}
		if n.High != nil {
			NormalizePos(n.High)
		}
		if n.Max != nil {
			NormalizePos(n.Max)
		}
		normalize(&n.Rbrack)

	case *ast.StarExpr:
		normalize(&n.Star)
		if n.X != nil {
			NormalizePos(n.X)
		}

	case *ast.StructType:
		normalize(&n.Struct)
		if n.Fields != nil {
			NormalizePos(n.Fields)
		}

	case *ast.SwitchStmt:
		normalize(&n.Switch)
		if n.Init != nil {
			NormalizePos(n.Init)
		}
		if n.Tag != nil {
			NormalizePos(n.Tag)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.TypeAssertExpr:
		if n.X != nil {
			NormalizePos(n.X)
		}
		normalize(&n.Lparen)
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		normalize(&n.Rparen)

	case *ast.TypeSpec:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		if n.Name != nil {
			NormalizePos(n.Name)
		}
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		if n.Comment != nil {
			NormalizePos(n.Comment)
		}

	case *ast.TypeSwitchStmt:
		normalize(&n.Switch)
		if n.Init != nil {
			NormalizePos(n.Init)
		}
		if n.Assign != nil {
			NormalizePos(n.Assign)
		}
		if n.Body != nil {
			NormalizePos(n.Body)
		}

	case *ast.UnaryExpr:
		normalize(&n.OpPos)
		if n.X != nil {
			NormalizePos(n.X)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			NormalizePos(n.Doc)
		}
		for _, x := range n.Names {
			NormalizePos(x)
		}
		if n.Type != nil {
			NormalizePos(n.Type)
		}
		for _, x := range n.Values {
			NormalizePos(x)
		}
		if n.Comment != nil {
			NormalizePos(n.Comment)
		}
	}
}
