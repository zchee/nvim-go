// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/rogpeppe/godef/go/ast"
	"github.com/rogpeppe/godef/go/parser"
	"github.com/rogpeppe/godef/go/printer"
	"github.com/rogpeppe/godef/go/token"
	"github.com/rogpeppe/godef/go/types"
)

var (
	defFiler         = "go#def#filer"
	vDefFiler        interface{}
	defFilerMode     = "go#def#filer_mode"
	vDefFilerMode    interface{}
	defNomodifiable  = "go#def#nomodifiable"
	vDefNomodifiable interface{}
	defDebug         = "go#def#debug"
	vDefDebug        interface{}
)

func init() {
	plugin.HandleFunction("GoDef", &plugin.FunctionOptions{}, funcDef)
}

func funcDef(v *vim.Vim, args []string) {
	go Def(v, args[0])
}

// Def definition to current cursor word.
// DEPRECATED: godef no longer mantained.
func Def(v *vim.Vim, file string) error {
	defer profile.Start(time.Now(), "GoDef")
	dir, _ := filepath.Split(file)
	var ctxt = context.Build{}
	defer ctxt.SetContext(dir)()

	gopath := strings.Split(build.Default.GOPATH, ":")
	for i, d := range gopath {
		gopath[i] = filepath.Join(d, "src")
	}
	types.GoPath = gopath

	v.Var(defDebug, &vDefDebug)
	if vDefDebug == int64(1) {
		types.Debug = true
	}

	v.Var(defNomodifiable, &vDefNomodifiable)

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	buf, err := v.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}
	src := bytes.Join(buf, []byte{'\n'})

	searchpos, err := buffer.ByteOffset(p)
	if err != nil {
		return v.WriteErr("cannot get current buffer byte offset")
	}

	pkgScope := ast.NewScope(parser.Universe)
	f, err := parser.ParseFile(types.FileSet, file, src, 0, pkgScope)
	if f == nil {
		nvim.Echomsg(v, "Godef: cannot parse %s: %v", file, err)
	}

	o := findIdentifier(v, f, searchpos)

	switch e := o.(type) {

	case *ast.ImportSpec:
		path := importPath(v, e)
		pkg, err := build.Default.Import(path, "", build.FindOnly)
		if err != nil {
			nvim.Echomsg(v, "Godef: error finding import path for %s: %s", path, err)
		}

		v.Var(defFilerMode, &vDefFilerMode)
		if vDefFilerMode.(string) != "" {
			v.Command(vDefFilerMode.(string))
		}
		v.Var(defFiler, &vDefFiler)
		return v.Command(vDefFiler.(string) + " " + pkg.Dir)

	case ast.Expr:
		if err := parseLocalPackage(file, f, pkgScope); err != nil {
			nvim.Echomsg(v, "Godef: error parseLocalPackage %v", err)
		}
		obj, _ := types.ExprType(e, types.DefaultImporter)
		if obj != nil {
			pos := types.FileSet.Position(types.DeclPos(obj))
			var loclist []*quickfix.ErrorlistData
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: pos.Filename,
				LNum:     pos.Line,
				Col:      pos.Column,
				Text:     pos.Filename,
			})
			if err := quickfix.SetLoclist(p, loclist); err != nil {
				nvim.Echomsg(v, "Godef: %s", err)
			}

			p.Command("silent! ll 1")
			if vDefNomodifiable == int64(1) {
				cb, err := v.CurrentBuffer()
				if err != nil {
					return nvim.Echoerr(v, "GoDef: %v", err)
				}
				p.SetBufferOption(cb, "modifiable", false)
			}
			p.FeedKeys("zz", "n", false)

			return p.Wait()
		}
		nvim.Echomsg(v, "Godef: not found of obj")

	default:
		nvim.Echomsg(v, "Godef: no declaration found for %v", pretty{e})
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
		typ = typ.Underlying(false, types.DefaultImporter)
		return fmt.Sprintf("type %s %v", obj.Name, prettyType{typ})
	}
	return fmt.Sprintf("unknown %s %v", obj.Name, typ.Kind)
}

type orderedObjects []*ast.Object

func (o orderedObjects) Less(i, j int) bool { return o[i].Name < o[j].Name }
func (o orderedObjects) Len() int           { return len(o) }
func (o orderedObjects) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func importPath(v *vim.Vim, n *ast.ImportSpec) string {
	p, err := strconv.Unquote(n.Path.Value)
	if err != nil {
		nvim.Echomsg(v, "Godef: invalid string literal %q in ast.ImportSpec", n.Path.Value)
	}
	return p
}

// findIdentifier looks for an identifier at byte-offset searchpos
// inside the parsed source represented by node.
// If it is part of a selector expression, it returns
// that expression rather than the identifier itself.
//
// As a special case, if it finds an import spec, it returns ImportSpec.
func findIdentifier(v *vim.Vim, f *ast.File, searchpos int) ast.Node {
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
		ast.Walk(FVisitor(visit), f)
		ec <- nil
	}()
	ev := <-ec
	if ev == nil {
		nvim.Echomsg(v, "Godef: no identifier found")
	}
	return ev
}

func parseExpr(v *vim.Vim, s *ast.Scope, expr string) ast.Expr {
	n, err := parser.ParseExpr(types.FileSet, "<arg>", expr, s)
	if err != nil {
		nvim.Echomsg(v, "Godef: cannot parse expression: %v", err)
	}
	switch n := n.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return n
	}
	nvim.Echomsg(v, "Godef: no identifier found in expression")
	return nil
}

// FVisitor for ast.Visit type.
type FVisitor func(n ast.Node) bool

// Visit for ast.Visit functions.
func (f FVisitor) Visit(n ast.Node) ast.Visitor {
	if f(n) {
		return f
	}
	return nil
}

// var errNoPkgFiles = errors.New("no more package files found")

// parseLocalPackage reads and parses all go files from the
// current directory that implement the same package name
// the principal source file, except the original source file
// itself, which will already have been parsed.
//
func parseLocalPackage(filename string, src *ast.File, pkgScope *ast.Scope) error {
	pkg := &ast.Package{src.Name.Name, pkgScope, nil, map[string]*ast.File{filename: src}}
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
		src, err := parser.ParseFile(types.FileSet, file, nil, 0, pkg.Scope)
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
	prog, _ := parser.ParseFile(types.FileSet, filename, nil, parser.PackageClauseOnly, nil)
	if prog != nil {
		return prog.Name.Name
	}
	return ""
}

func hasSuffix(s, suff string) bool {
	return len(s) >= len(suff) && s[len(s)-len(suff):] == suff
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
