// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gb

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

func FindGbProject(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("project root is blank")
	}
	start := path
	for path != filepath.Dir(path) {
		root := filepath.Join(path, "src")
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				path = filepath.Dir(path)
				continue
			}
			return "", err
		}
		return path, nil
	}

	return "", fmt.Errorf(`could not find project root in "%s" or its parents`, start)
}

// GoPath returns a GOPATH for path p.
func GoPath(p string) string {
	goPath := os.Getenv("GOPATH")
	r := runtime.GOROOT()
	if r != "" {
		goPath = goPath + string(filepath.ListSeparator) + r
	}

	p = filepath.Clean(p)

	for _, root := range filepath.SplitList(goPath) {
		if strings.HasPrefix(p, filepath.Join(root, "src")+string(filepath.Separator)) {
			return goPath
		}
	}

	project, err := FindGbProject(p)
	if err == nil {
		parent, child := filepath.Split(project)
		if child == "vendor" {
			project = parent[:len(parent)-1]
		}
		return project + string(filepath.ListSeparator) +
			filepath.Join(project, "vendor") + string(filepath.ListSeparator) + goPath
	}

	return goPath
}

var goBuildDefaultMu sync.Mutex

// WithGoBuildForPath sets the go/build Default.GOPATH to GoPath(p) under a
// mutex. The returned function restores Default.GOPATH to its original value
// and unlocks the mutex.
//
// This function intended to be used to the golang.org/x/tools/imports and
// other packages that use go/build Default.
func WithGoBuildForPath(p string) func() {
	goBuildDefaultMu.Lock()
	original := build.Default.GOPATH
	build.Default.GOPATH = GoPath(p)
	os.Setenv("GOPATH", build.Default.GOPATH)
	build.Default.UseAllFiles = false
	return func() {
		build.Default.GOPATH = original
		goBuildDefaultMu.Unlock()
	}
}

// Package represents a Go package.
type Package struct {
	FSet     *token.FileSet
	Build    *build.Package
	AST      *ast.Package
	Doc      *doc.Package
	Examples []*doc.Example
	Errors   []error
}

// Flags for LoadPackage.
const (
	LoadDoc = 1 << iota
	LoadExamples
	LoadUnexported
)

// LoadPackage Import returns details about the Go package named by the import
// path, interpreting local import paths relative to the srcDir directory.
func LoadPackage(importPath string, srcDir string, flags int) (*Package, error) {
	bpkg, err := build.Default.Import(importPath, srcDir, build.ImportComment)
	if _, ok := err.(*build.NoGoError); ok {
		return &Package{Build: bpkg}, nil
	}
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		FSet:  token.NewFileSet(),
		Build: bpkg,
	}

	files := make(map[string]*ast.File)
	for _, name := range append(pkg.Build.GoFiles, pkg.Build.CgoFiles...) {
		file, err := pkg.parseFile(name)
		if err != nil {
			pkg.Errors = append(pkg.Errors, err)
			continue
		}
		files[name] = file
	}

	pkg.AST, _ = ast.NewPackage(pkg.FSet, files, simpleImporter, nil)

	if flags&LoadDoc != 0 {
		mode := doc.Mode(0)
		if pkg.Build.ImportPath == "builtin" || flags&LoadUnexported != 0 {
			mode |= doc.AllDecls
		}
		pkg.Doc = doc.New(pkg.AST, pkg.Build.ImportPath, mode)
		if pkg.Build.ImportPath == "builtin" {
			for _, t := range pkg.Doc.Types {
				pkg.Doc.Funcs = append(pkg.Doc.Funcs, t.Funcs...)
				t.Funcs = nil
			}
			sort.Sort(byFuncName(pkg.Doc.Funcs))
		}
	}

	if flags&LoadExamples != 0 {
		for _, name := range append(pkg.Build.TestGoFiles, pkg.Build.XTestGoFiles...) {
			file, err := pkg.parseFile(name)
			if err != nil {
				pkg.Errors = append(pkg.Errors, err)
				continue
			}
			pkg.Examples = append(pkg.Examples, doc.Examples(file)...)
		}
	}

	return pkg, nil
}

type byFuncName []*doc.Func

func (s byFuncName) Len() int           { return len(s) }
func (s byFuncName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byFuncName) Less(i, j int) bool { return s[i].Name < s[j].Name }

func (pkg *Package) parseFile(name string) (*ast.File, error) {
	p, err := ioutil.ReadFile(filepath.Join(pkg.Build.Dir, name))
	if err != nil {
		return nil, err
	}
	// overwrite //line comments
	for _, m := range linePat.FindAllIndex(p, -1) {
		for i := m[0] + 2; i < m[1]; i++ {
			p[i] = ' '
		}
	}
	return parser.ParseFile(pkg.FSet, name, p, parser.ParseComments)
}

var linePat = regexp.MustCompile(`(?m)^//line .*$`)

func simpleImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
	pkg := imports[path]
	if pkg != nil {
		return pkg, nil
	}

	n := GuessPackageNameFromPath(path)
	if n == "" {
		return nil, errors.New("package not found")
	}

	pkg = ast.NewObj(ast.Pkg, n)
	pkg.Data = ast.NewScope(nil)
	imports[path] = pkg
	return pkg, nil
}

var packageNamePats = []*regexp.Regexp{
	// Last element with .suffix removed.
	regexp.MustCompile(`/([^-./]+)[-.](?:git|svn|hg|bzr|v\d+)$`),

	// Last element with "go" prefix or suffix removed.
	regexp.MustCompile(`/([^-./]+)[-.]go$`),
	regexp.MustCompile(`/go[-.]([^-./]+)$`),

	// It's also common for the last element of the path to contain an
	// extra "go" prefix, but not always. TODO: examine unresolved ids to
	// detect when trimming the "go" prefix is appropriate.

	// Last component of path.
	regexp.MustCompile(`([^/]+)$`),
}

// GuessPackageNameFromPath guesses the package name from the package path.
func GuessPackageNameFromPath(path string) string {
	// Guess the package name without importing it.
	for _, pat := range packageNamePats {
		m := pat.FindStringSubmatch(path)
		if m != nil {
			return m[1]
		}
	}
	return ""
}
