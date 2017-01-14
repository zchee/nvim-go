// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
	"github.com/rogpeppe/godef/go/ast"
	"github.com/rogpeppe/godef/go/parser"
	"github.com/rogpeppe/godef/go/printer"
	"github.com/rogpeppe/godef/go/token"
	"github.com/rogpeppe/godef/go/types"
)

func (c *Commands) cmdDef(file string) {
	go c.Def(file)
}

// Def definition to current cursor word.
// DEPRECATED: godef no longer mantained.
func (c *Commands) Def(file string) error {
	defer nvimutil.Profile(time.Now(), "GoDef")
	defer c.ctx.SetContext(filepath.Dir(file))()

	// types.Debug = true

	var (
		b nvim.Buffer
		w nvim.Window
	)
	c.Pipeline.CurrentBuffer(&b)
	c.Pipeline.CurrentWindow(&w)
	if err := c.Pipeline.Wait(); err != nil {
		return err
	}

	buf, err := c.Nvim.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}
	src := bytes.Join(buf, []byte{'\n'})

	searchpos, err := nvimutil.ByteOffsetPipe(c.Pipeline, b, w)
	if err != nil {
		return c.Nvim.WriteErr("cannot get current buffer byte offset")
	}

	pkgScope := ast.NewScope(parser.Universe)
	f, err := parser.ParseFile(types.FileSet, file, src, 0, pkgScope, types.DefaultImportPathToName)
	if f == nil {
		nvimutil.Echomsg(c.Nvim, "Godef: cannot parse %s: %v", file, err)
	}

	o := findIdentifier(c.Nvim, f, searchpos)

	switch e := o.(type) {
	case ast.Expr:
		if err := parseLocalPackage(file, f, pkgScope); err != nil {
			nvimutil.Echomsg(c.Nvim, "Godef: error parseLocalPackage %v", err)
		}
		obj, _ := types.ExprType(e, types.DefaultImporter, types.FileSet)
		if obj != nil {
			pos := types.FileSet.Position(types.DeclPos(obj))
			var loclist []*nvim.QuickfixError
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: pos.Filename,
				LNum:     pos.Line,
				Col:      pos.Column,
				Text:     pos.Filename,
			})
			if err := nvimutil.SetLoclist(c.Nvim, loclist); err != nil {
				nvimutil.Echomsg(c.Nvim, "Godef: %s", err)
			}

			c.Pipeline.Command(fmt.Sprintf("edit %s", pos.Filename))
			c.Pipeline.SetWindowCursor(w, [2]int{pos.Line, pos.Column - 1})
			c.Pipeline.Command("normal zz")

			return nil
		}
		nvimutil.Echomsg(c.Nvim, "Godef: not found of obj")

	default:
		nvimutil.Echomsg(c.Nvim, "Godef: no declaration found for %v", pretty{e})
	}
	return nil
}

func typeStrMap(obj *ast.Object, typ types.Type) map[string]string {
	switch obj.Kind {
	case ast.Fun, ast.Var:
		dict := map[string]string{
			"Object.Kind":   obj.Kind.String(),
			"Object.Name":   obj.Name,
			"Type.Kind":     typ.Kind.String(),
			"Type.Pkg":      typ.Pkg,
			"Type.String()": typ.String(),
			// "Object.Decl": obj.Decl,
			// "Object.Data":   obj.Data,
			// "Object.Type":   obj.Type,
			// "Object.Pos()":  obj.Pos(),
			// "Type.Node":     typ.Node,
		}
		return dict
		// 	return fmt.Sprintf("%s %v", typ.obj.Name, prettyType{typ})
		// case ast.Pkg:
		// 	return fmt.Sprintf("import (%s %s)", obj.Name, typ.Node.(*ast.ImportSpec).Path.Value)
		// case ast.Con:
		// 	if decl, ok := obj.Decl.(*ast.ValueSpec); ok {
		// 		return fmt.Sprintf("const %s %v = %s", obj.Name, prettyType{typ}, pretty{decl.Values[0]})
		// 	}
		// 	return fmt.Sprintf("const %s %v", obj.Name, prettyType{typ})
		// case ast.Lbl:
		// 	return fmt.Sprintf("label %s", obj.Name)
		// case ast.Typ:
		// 	typ = typ.Underlying(false, types.DefaultImporter)
		// 	return fmt.Sprintf("type %s %v", obj.Name, prettyType{typ})
		// }
		// return fmt.Sprintf("unknown %s %v", obj.Name, typ.Kind)
	}
	return map[string]string{}
}

func typeStr(obj *ast.Object, typ types.Type) string {
	switch obj.Kind {
	case ast.Fun, ast.Var:
		return fmt.Sprintf("%s %v", obj.Name, prettyType{typ})
	case ast.Pkg:
		return fmt.Sprintf("import (%s %s)", obj.Name, typ.Node.(*ast.ImportSpec).Path.Value)
	case ast.Con:
		if decl, ok := obj.Decl.(*ast.ValueSpec); ok {
			return fmt.Sprintf("const %s %v = %s", obj.Name, prettyType{typ}, pretty{decl.Values[0]})
		}
		return fmt.Sprintf("const %s %v", obj.Name, prettyType{typ})
	case ast.Lbl:
		return fmt.Sprintf("label %s", obj.Name)
	case ast.Typ:
		typ = typ.Underlying(false)
		return fmt.Sprintf("type %s %v", obj.Name, prettyType{typ})
	}
	return fmt.Sprintf("unknown %s %v", obj.Name, typ.Kind)
}

type orderedObjects []*ast.Object

func (o orderedObjects) Less(i, j int) bool { return o[i].Name < o[j].Name }
func (o orderedObjects) Len() int           { return len(o) }
func (o orderedObjects) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func importPath(v *nvim.Nvim, n *ast.ImportSpec) string {
	p, err := strconv.Unquote(n.Path.Value)
	if err != nil {
		nvimutil.Echomsg(v, "Godef: invalid string literal %q in ast.ImportSpec", n.Path.Value)
	}
	return p
}

// findIdentifier looks for an identifier at byte-offset searchpos
// inside the parsed source represented by node.
// If it is part of a selector expression, it returns
// that expression rather than the identifier itself.
//
// As a special case, if it finds an import spec, it returns ImportSpec.
func findIdentifier(v *nvim.Nvim, f *ast.File, searchpos int) ast.Node {
	ec := make(chan ast.Node)
	found := func(startPos, endPos token.Pos) bool {
		start := types.FileSet.Position(startPos).Offset
		end := start + int(endPos-startPos)
		return start <= searchpos && searchpos <= end
	}
	go func() {
		var visit func(ast.Node) bool
		visit = func(n ast.Node) bool {
			var startPos token.Pos
			switch n := n.(type) {
			default:
				return true
			case *ast.Ident:
				startPos = n.NamePos
			case *ast.SelectorExpr:
				startPos = n.Sel.NamePos
			case *ast.ImportSpec:
				startPos = n.Pos()
			case *ast.StructType:
				// If we find an anonymous bare field in a
				// struct type, its definition points to itself,
				// but we actually want to go elsewhere,
				// so assume (dubiously) that the expression
				// works globally and return a new node for it.
				for _, field := range n.Fields.List {
					if field.Names != nil {
						continue
					}
					t := field.Type
					if pt, ok := field.Type.(*ast.StarExpr); ok {
						t = pt.X
					}
					if id, ok := t.(*ast.Ident); ok {
						if found(id.NamePos, id.End()) {
							ec <- parseExpr(v, f.Scope, id.Name)
							runtime.Goexit()
						}
					}
				}
				return true
			}
			if found(startPos, n.End()) {
				ec <- n
				runtime.Goexit()
			}
			return true
		}
		ast.Walk(defVisitor(visit), f)
		ec <- nil
	}()
	ev := <-ec
	if ev == nil {
		nvimutil.Echomsg(v, "Godef: no identifier found")
	}
	return ev
}

func parseExpr(v *nvim.Nvim, s *ast.Scope, expr string) ast.Expr {
	n, err := parser.ParseExpr(types.FileSet, "<arg>", expr, s, types.DefaultImportPathToName)
	if err != nil {
		nvimutil.Echomsg(v, "Godef: cannot parse expression: %v", err)
	}
	switch n := n.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return n
	}
	nvimutil.Echomsg(v, "Godef: no identifier found in expression")
	return nil
}

// defVisitor for ast.Visit type.
type defVisitor func(n ast.Node) bool

// Visit for ast.Visit functions.
func (f defVisitor) Visit(n ast.Node) ast.Visitor {
	if f(n) {
		return f
	}
	return nil
}

// parseLocalPackage reads and parses all go files from the
// current directory that implement the same package name
// the principal source file, except the original source file
// itself, which will already have been parsed.
//
func parseLocalPackage(filename string, src *ast.File, pkgScope *ast.Scope) error {
	pkg := &ast.Package{
		Name:    src.Name.Name,
		Scope:   pkgScope,
		Imports: nil,
		Files:   map[string]*ast.File{filename: src},
	}
	d, f := filepath.Split(filename)
	if d == "" {
		d = "./"
	}
	fd, err := os.Open(d)
	if err != nil {
		return nil
	}
	defer fd.Close()

	list, err := fd.Readdirnames(-1)
	if err != nil {
		return nil
	}

	for _, pf := range list {
		file := filepath.Join(d, pf)
		if !strings.HasSuffix(pf, ".go") ||
			pf == f ||
			pkgName(file) != pkg.Name {
			continue
		}
		src, err := parser.ParseFile(types.FileSet, file, nil, 0, pkg.Scope, types.DefaultImportPathToName)
		if err == nil {
			pkg.Files[file] = src
		}
	}
	if len(pkg.Files) == 1 {
		return nil
	}
	return nil
}

// pkgName returns the package name implemented by the go source filename
func pkgName(filename string) string {
	prog, _ := parser.ParseFile(types.FileSet, filename, nil, parser.PackageClauseOnly, nil, types.DefaultImportPathToName)
	if prog != nil {
		return prog.Name.Name
	}
	return ""
}

type pretty struct {
	n interface{}
}

func (p pretty) String() string {
	var b bytes.Buffer
	printer.Fprint(&b, types.FileSet, p.n)
	return b.String()
}

type prettyType struct {
	n types.Type
}

func (p prettyType) String() string {
	// TODO print path package when appropriate.
	// Current issues with using p.n.Pkg:
	//	- we should actually print the local package identifier
	//	rather than the package path when possible.
	//	- p.n.Pkg is non-empty even when
	//	the type is not relative to the package.
	return pretty{p.n.Node}.String()
}
