// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"nvim-go/internal/guru"

	"github.com/pkg/errors"
	"golang.org/x/tools/cmd/guru/serial"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
)

// Definition parse definition from the current cursor.
func Definition(q *guru.Query) (*serial.Definition, error) {
	// First try the simple resolution done by parser.
	// It only works for intra-file references but it is very fast.
	// (Extending this approach to all the files of the package,
	// resolved using ast.NewPackage, was not worth the effort.)
	{
		qpos, err := fastQueryPos(q.Build, q.Pos)
		if err != nil {
			return nil, err
		}

		id, _ := qpos.path[0].(*ast.Ident)
		if id == nil {
			err := errors.New("no identifier here")
			return nil, err
		}

		// Did the parser resolve it to a local object?
		if obj := id.Obj; obj != nil && obj.Pos().IsValid() {
			return &serial.Definition{
				ObjPos: qpos.fset.Position(obj.Pos()).String(),
				Desc:   fmt.Sprintf("%s %s", obj.Kind, obj.Name),
			}, nil
		}

		// Qualified identifier?
		if pkg := guru.PackageForQualIdent(qpos.path, id); pkg != "" {
			srcdir := filepath.Dir(qpos.fset.File(qpos.start).Name())
			tok, pos, err := guru.FindPackageMember(q.Build, qpos.fset, srcdir, pkg, id.Name)
			if err != nil {
				return nil, err
			}
			return &serial.Definition{
				ObjPos: qpos.fset.Position(pos).String(),
				Desc:   fmt.Sprintf("%s %s.%s", tok, pkg, id.Name),
			}, nil
		}

		// Fall back on the type checker.
	}

	// Set loader.Config, same allowErrors() function result except CgoEnabled = false
	q.Build.CgoEnabled = false
	// Run the type checker.
	lconf := loader.Config{
		Build:       q.Build,
		AllowErrors: true,
		// AllErrors makes the parser always return an AST instead of
		// bailing out after 10 errors and returning an empty ast.File.
		ParserMode: parser.AllErrors,
		TypeChecker: types.Config{
			Error: func(err error) {},
		},
	}

	if _, err := importQueryPackage(q.Pos, &lconf); err != nil {
		return nil, err
	}

	// Load/parse/type-check the program.
	lprog, err := lconf.Load()
	if err != nil {
		return nil, err
	}

	qpos, err := parseQueryPos(lprog, q.Pos, false)
	if err != nil {
		return nil, err
	}

	id, _ := qpos.path[0].(*ast.Ident)
	if id == nil {
		err := errors.New("no identifier here")
		return nil, err
	}

	obj := qpos.info.ObjectOf(id)
	if obj == nil {
		// Happens for y in "switch y := x.(type)",
		// and the package declaration,
		// but I think that's all.
		err := errors.New("no object for identifier")
		return nil, err
	}

	if !obj.Pos().IsValid() {
		err := errors.Errorf("%s is built in", obj.Name())
		return nil, err
	}

	return &serial.Definition{
		ObjPos: qpos.fset.Position(obj.Pos()).String(),
		Desc:   qpos.ObjectString(obj),
	}, nil
}

// A QueryPos represents the position provided as input to a query:
// a textual extent in the program's source code, the AST node it
// corresponds to, and the package to which it belongs.
// Instances are created by parseQueryPos.
type queryPos struct {
	fset       *token.FileSet
	start, end token.Pos           // source extent of query
	path       []ast.Node          // AST path from query node to root of ast.File
	exact      bool                // 2nd result of PathEnclosingInterval
	info       *loader.PackageInfo // type info for the queried package (nil for fastQueryPos)
}

// TypeString prints type T relative to the query position.
func (qpos *queryPos) typeString(T types.Type) string {
	return types.TypeString(T, types.RelativeTo(qpos.info.Pkg))
}

// ObjectString prints object obj relative to the query position.
func (qpos *queryPos) ObjectString(obj types.Object) string {
	return types.ObjectString(obj, types.RelativeTo(qpos.info.Pkg))
}

// SelectionString prints selection sel relative to the query position.
func (qpos *queryPos) selectionString(sel *types.Selection) string {
	return types.SelectionString(sel, types.RelativeTo(qpos.info.Pkg))
}

// parseOctothorpDecimal returns the numeric value if s matches "#%d",
// otherwise -1.
func parseOctothorpDecimal(s string) int {
	if s != "" && s[0] == '#' {
		if s, err := strconv.ParseInt(s[1:], 10, 32); err == nil {
			return int(s)
		}
	}
	return -1
}

// parsePos parses a string of the form "file:pos" or
// file:start,end" where pos, start, end match #%d and represent byte
// offsets, and returns its components.
//
// (Numbers without a '#' prefix are reserved for future use,
// e.g. to indicate line/column positions.)
//
func parsePos(pos string) (filename string, startOffset, endOffset int, err error) {
	if pos == "" {
		err = fmt.Errorf("no source position specified")
		return
	}

	colon := strings.LastIndex(pos, ":")
	if colon < 0 {
		err = fmt.Errorf("bad position syntax %q", pos)
		return
	}
	filename, offset := pos[:colon], pos[colon+1:]
	startOffset = -1
	endOffset = -1
	if hyphen := strings.Index(offset, ","); hyphen < 0 {
		// e.g. "foo.go:#123"
		startOffset = parseOctothorpDecimal(offset)
		endOffset = startOffset
	} else {
		// e.g. "foo.go:#123,#456"
		startOffset = parseOctothorpDecimal(offset[:hyphen])
		endOffset = parseOctothorpDecimal(offset[hyphen+1:])
	}
	if startOffset < 0 || endOffset < 0 {
		err = fmt.Errorf("invalid offset %q in query position", offset)
		return
	}
	return
}

// fileOffsetToPos translates the specified file-relative byte offsets
// into token.Pos form.  It returns an error if the file was not found
// or the offsets were out of bounds.
//
func fileOffsetToPos(file *token.File, startOffset, endOffset int) (start, end token.Pos, err error) {
	// Range check [start..end], inclusive of both end-points.

	if 0 <= startOffset && startOffset <= file.Size() {
		start = file.Pos(int(startOffset))
	} else {
		err = fmt.Errorf("start position is beyond end of file")
		return
	}

	if 0 <= endOffset && endOffset <= file.Size() {
		end = file.Pos(int(endOffset))
	} else {
		err = fmt.Errorf("end position is beyond end of file")
		return
	}

	return
}

// sameFile returns true if x and y have the same basename and denote
// the same file.
//
func sameFile(x, y string) bool {
	if filepath.Base(x) == filepath.Base(y) { // (optimisation)
		if xi, err := os.Stat(x); err == nil {
			if yi, err := os.Stat(y); err == nil {
				return os.SameFile(xi, yi)
			}
		}
	}
	return false
}

// fastQueryPos parses the position string and returns a queryPos.
// It parses only a single file and does not run the type checker.
func fastQueryPos(ctxt *build.Context, pos string) (*queryPos, error) {
	filename, startOffset, endOffset, err := parsePos(pos)
	if err != nil {
		return nil, err
	}

	// Parse the file, opening it the file via the build.Context
	// so that we observe the effects of the -modified flag.
	fset := token.NewFileSet()
	fset.AddFile(filename, fset.Base(), startOffset)
	// cwd, _ := os.Getwd()
	cwd, _ := filepath.Split(filename)
	f, err := buildutil.ParseFile(fset, ctxt, nil, filepath.Clean(cwd), filename, parser.Mode(0))
	// ParseFile usually returns a partial file along with an error.
	// Only fail if there is no file.
	if f == nil {
		return nil, err
	}
	if !f.Pos().IsValid() {
		return nil, fmt.Errorf("%s is not a Go source file", filename)
	}

	start, end, err := fileOffsetToPos(fset.File(f.Pos()), startOffset, endOffset)
	if err != nil {
		return nil, err
	}

	path, exact := astutil.PathEnclosingInterval(f, start, end)
	if path == nil {
		return nil, fmt.Errorf("no syntax here")
	}

	return &queryPos{fset, start, end, path, exact, nil}, nil
}

// ParseQueryPos parses the source query position pos and returns the
// AST node of the loaded program lprog that it identifies.
// If needExact, it must identify a single AST subtree;
// this is appropriate for queries that allow fairly arbitrary syntax,
// e.g. "describe".
//
func parseQueryPos(lprog *loader.Program, pos string, needExact bool) (*queryPos, error) {
	filename, startOffset, endOffset, err := parsePos(pos)
	if err != nil {
		return nil, err
	}

	// Find the named file among those in the loaded program.
	var file *token.File
	lprog.Fset.Iterate(func(f *token.File) bool {
		if sameFile(filename, f.Name()) {
			file = f
			return false // done
		}
		return true // continue
	})
	if file == nil {
		return nil, fmt.Errorf("file %s not found in loaded program", filename)
	}

	start, end, err := fileOffsetToPos(file, startOffset, endOffset)
	if err != nil {
		return nil, err
	}
	info, path, exact := lprog.PathEnclosingInterval(start, end)
	if path == nil {
		return nil, fmt.Errorf("no syntax here")
	}
	if needExact && !exact {
		return nil, fmt.Errorf("ambiguous selection within %s", astutil.NodeDescription(path[0]))
	}
	return &queryPos{lprog.Fset, start, end, path, exact, info}, nil
}

// importQueryPackage finds the package P containing the
// query position and tells conf to import it.
// It returns the package's path.
func importQueryPackage(pos string, conf *loader.Config) (string, error) {
	fqpos, err := fastQueryPos(conf.Build, pos)
	if err != nil {
		return "", err // bad query
	}
	filename := fqpos.fset.File(fqpos.start).Name()

	_, importPath, err := guru.GuessImportPath(filename, conf.Build)
	if err != nil {
		return "", err // can't find GOPATH dir
	}

	// Check that it's possible to load the queried package.
	// (e.g. guru tests contain different 'package' decls in same dir.)
	// Keep consistent with logic in loader/util.go!
	cfg2 := *conf.Build
	cfg2.CgoEnabled = false
	bp, err := cfg2.Import(importPath, "", 0)
	if err != nil {
		return "", err // no files for package
	}

	switch pkgContainsFile(bp, filename) {
	case 'T':
		conf.ImportWithTests(importPath)
	case 'X':
		conf.ImportWithTests(importPath)
		importPath += "_test" // for TypeCheckFuncBodies
	case 'G':
		conf.Import(importPath)
	default:
		// This happens for ad-hoc packages like
		// $GOROOT/src/net/http/triv.go.
		return "", fmt.Errorf("package %q doesn't contain file %s",
			importPath, filename)
	}

	conf.TypeCheckFuncBodies = func(p string) bool { return p == importPath }

	return importPath, nil
}

// pkgContainsFile reports whether file was among the packages Go
// files, Test files, eXternal test files, or not found.
func pkgContainsFile(bp *build.Package, filename string) byte {
	for i, files := range [][]string{bp.GoFiles, bp.TestGoFiles, bp.XTestGoFiles} {
		for _, file := range files {
			if sameFile(filepath.Join(bp.Dir, file), filename) {
				return "GTX"[i]
			}
		}
	}
	return 0 // not found
}
