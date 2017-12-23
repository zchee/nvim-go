// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"go/build"
	"io/ioutil"
	"os"
	pathPkg "path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	"github.com/golang/lint"
	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

func (c *Command) cmdLint(v *nvim.Nvim, args []string, file string) {
	// Cleanup error list
	delete(c.buildctxt.Errlist, "Lint")

	go func() {
		errlist, err := c.Lint(args, file)
		if err != nil {
			nvimutil.ErrorWrap(c.Nvim, err)
		}
		c.buildctxt.Errlist["Lint"] = errlist
		nvimutil.ErrorList(c.Nvim, c.buildctxt.Errlist, true)
	}()
}

type lintMode string

const (
	current lintMode = "current"
	root    lintMode = "root"
)

// Lint lints a go source file. The argument is a filename or directory path.
// TODO(zchee): Support go packages.
func (c *Command) Lint(args []string, file string) ([]*nvim.QuickfixError, error) {
	defer nvimutil.Profile(time.Now(), "GoLint")

	var (
		errlist []*nvim.QuickfixError
		err     error
	)
	switch {
	case len(args) == 0:
		switch lintMode(config.GolintMode) {
		case current:
			errlist, err = c.lintDir(filepath.Dir(file))
		case root:
			var rootDir string
			switch c.buildctxt.Build.Tool {
			case "go":
				root, err := pathutil.PackageID(c.buildctxt.Build.ProjectRoot)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				rootDir = root
			case "gb":
				rootDir = filepath.Base(c.buildctxt.Build.ProjectRoot)
			}
			for _, pkgname := range importPaths([]string{rootDir + "/..."}) {
				errors, err := c.lintPackage(pkgname)
				if err != nil {
					return nil, err
				}
				errlist = append(errlist, errors...)
			}
		}

	case len(args) == 1:
		path := args[0]
		if path == "%" {
			path = file
		}
		switch {
		case pathutil.IsDir(path):
			errlist, err = c.lintDir(path)
		case pathutil.IsExist(path):
			errlist, err = c.lintFiles(path)
		case !pathutil.IsDir(path) || !pathutil.IsExist(path):
			for _, pkgname := range importPaths(args) {
				errlist, err = c.lintPackage(pkgname)
			}
		}

	case len(args) >= 2:
		errlist, err = c.lintFiles(args...)
	}

	return errlist, errors.WithStack(err)
}

// TODO(zchee): Support list of go packages.
func (c *Command) cmdLintComplete(a *nvim.CommandCompletionArgs, cwd string) (filelist []string, err error) {
	files, err := nvimutil.CompleteFiles(c.Nvim, a, cwd)
	if err != nil {
		return nil, err
	}
	filelist = append(filelist, files...)

	return filelist, nil
}

// ----------------------------------------------------------------------------
// The below code is based by github.com/golang/lint/golint/golint.go

func (c *Command) lintFiles(filenames ...string) ([]*nvim.QuickfixError, error) {
	files := make(map[string][]byte)
	for _, filename := range filenames {
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}
		files[filename] = src
	}

	l := new(lint.Linter)
	ps, err := l.LintFiles(files)
	if err != nil {
		return nil, nvimutil.ErrorWrap(c.Nvim, err)
	}

	var cwdRes interface{}
	if err := c.Nvim.Eval("getcwd()", &cwdRes); err != nil {
		return nil, nvimutil.ErrorWrap(c.Nvim, err)
	}
	cwd := cwdRes.(string)

	var errlist []*nvim.QuickfixError
	for _, p := range ps {
		if p.Confidence >= config.GolintMinConfidence {
			file := p.Position.Filename
			if contain(file, config.GolintIgnore) {
				continue
			}
			frel, err := filepath.Rel(cwd, file)
			if err == nil {
				file = frel
			}

			errlist = append(errlist, &nvim.QuickfixError{
				FileName: file,
				LNum:     p.Position.Line,
				Col:      p.Position.Column,
				Text:     p.Text,
			})
		}
	}

	return errlist, nil
}

func contain(s string, ignore []string) bool {
	for _, f := range ignore {
		if strings.Index(s, f) > 0 {
			return true
		}
	}
	return false
}

func (c *Command) lintDir(dirname string) ([]*nvim.QuickfixError, error) {
	pkg, err := build.ImportDir(dirname, 0)
	return c.lintImportedPackage(pkg, err)
}

func (c *Command) lintPackage(pkgname string) ([]*nvim.QuickfixError, error) {
	pkg, err := build.Import(pkgname, ".", 0)
	return c.lintImportedPackage(pkg, err)
}

func (c *Command) lintImportedPackage(pkg *build.Package, err error) ([]*nvim.QuickfixError, error) {
	if err != nil {
		if _, nogo := err.(*build.NoGoError); nogo {
			// Don't complain if the failure is due to no Go source files.
			return nil, nil
		}
		return nil, err
	}

	var files []string
	files = append(files, pkg.GoFiles...)
	files = append(files, pkg.CgoFiles...)
	files = append(files, pkg.TestGoFiles...)
	if pkg.Dir != "." {
		for i, f := range files {
			files[i] = filepath.Join(pkg.Dir, f)
		}
	}
	// TODO(dsymonds): Do foo_test too (pkg.XTestGoFiles)

	return c.lintFiles(files...)
}

// ----------------------------------------------------------------------------
// The below code is carried from github.com/golang/lint/golint/import.go

var (
	goroot    = filepath.Clean(runtime.GOROOT())
	gorootSrc = filepath.Join(goroot, "src")
)

// importPathsNoDotExpansion returns the import paths to use for the given
// command line, but it does no ... expansion.
func importPathsNoDotExpansion(args []string) []string {
	if len(args) == 0 {
		return []string{"."}
	}
	var out []string
	for _, a := range args {
		// Arguments are supposed to be import paths, but
		// as a courtesy to Windows developers, rewrite \ to /
		// in command-line arguments.  Handles .\... and so on.
		if filepath.Separator == '\\' {
			a = strings.Replace(a, `\`, `/`, -1)
		}

		// Put argument in canonical form, but preserve leading ./.
		if strings.HasPrefix(a, "./") {
			a = "./" + pathPkg.Clean(a)
			if a == "./." {
				a = "."
			}
		} else {
			a = pathPkg.Clean(a)
		}
		if a == "all" || a == "std" {
			out = append(out, allPackages(a)...)
			continue
		}
		out = append(out, a)
	}
	return out
}

// importPaths returns the import paths to use for the given command line.
func importPaths(args []string) []string {
	args = importPathsNoDotExpansion(args)
	var out []string
	for _, a := range args {
		if strings.Contains(a, "...") {
			if build.IsLocalImport(a) {
				out = append(out, allPackagesInFS(a)...)
			} else {
				out = append(out, allPackages(a)...)
			}
			continue
		}
		out = append(out, a)
	}
	return out
}

// matchPattern(pattern)(name) reports whether
// name matches pattern.  Pattern is a limited glob
// pattern in which '...' means 'any string' and there
// is no other special syntax.
func matchPattern(pattern string) func(name string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.Replace(re, `\.\.\.`, `.*`, -1)
	// Special case: foo/... matches foo too.
	if strings.HasSuffix(re, `/.*`) {
		re = re[:len(re)-len(`/.*`)] + `(/.*)?`
	}
	reg := regexp.MustCompile(`^` + re + `$`)
	return func(name string) bool {
		return reg.MatchString(name)
	}
}

// hasPathPrefix reports whether the path s begins with the
// elements in prefix.
func hasPathPrefix(s, prefix string) bool {
	switch {
	default:
		return false
	case len(s) == len(prefix):
		return s == prefix
	case len(s) > len(prefix):
		if prefix != "" && prefix[len(prefix)-1] == '/' {
			return strings.HasPrefix(s, prefix)
		}
		return s[len(prefix)] == '/' && s[:len(prefix)] == prefix
	}
}

// treeCanMatchPattern(pattern)(name) reports whether
// name or children of name can possibly match pattern.
// Pattern is the same limited glob accepted by matchPattern.
func treeCanMatchPattern(pattern string) func(name string) bool {
	wildCard := false
	if i := strings.Index(pattern, "..."); i >= 0 {
		wildCard = true
		pattern = pattern[:i]
	}
	return func(name string) bool {
		return len(name) <= len(pattern) && hasPathPrefix(pattern, name) ||
			wildCard && strings.HasPrefix(name, pattern)
	}
}

// allPackages returns all the packages that can be found
// under the $GOPATH directories and $GOROOT matching pattern.
// The pattern is either "all" (all packages), "std" (standard packages)
// or a path including "...".
func allPackages(pattern string) []string {
	pkgs := matchPackages(pattern)
	if len(pkgs) == 0 {
		// fmt.Fprintf(os.Stderr, "warning: %q matched no packages\n", pattern)
	}
	return pkgs
}

func matchPackages(pattern string) []string {
	match := func(string) bool { return true }
	treeCanMatch := func(string) bool { return true }
	if pattern != "all" && pattern != "std" {
		match = matchPattern(pattern)
		treeCanMatch = treeCanMatchPattern(pattern)
	}

	have := map[string]bool{
		"builtin": true, // ignore pseudo-package that exists only for documentation
	}
	buildContext := build.Default
	if !buildContext.CgoEnabled {
		have["runtime/cgo"] = true // ignore during walk
	}
	var pkgs []string

	// Commands
	cmd := filepath.Join(goroot, "src/cmd") + string(filepath.Separator)
	filepath.Walk(cmd, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() || path == cmd {
			return nil
		}
		name := path[len(cmd):]
		if !treeCanMatch(name) {
			return filepath.SkipDir
		}
		// Commands are all in cmd/, not in subdirectories.
		if strings.Contains(name, string(filepath.Separator)) {
			return filepath.SkipDir
		}

		// We use, e.g., cmd/gofmt as the pseudo import path for gofmt.
		name = "cmd/" + name
		if have[name] {
			return nil
		}
		have[name] = true
		if !match(name) {
			return nil
		}
		_, err = buildContext.ImportDir(path, 0)
		if err != nil {
			if _, noGo := err.(*build.NoGoError); !noGo {
				// log.Print(err)
			}
			return nil
		}
		pkgs = append(pkgs, name)
		return nil
	})

	for _, src := range buildContext.SrcDirs() {
		if (pattern == "std" || pattern == "cmd") && src != gorootSrc {
			continue
		}
		src = filepath.Clean(src) + string(filepath.Separator)
		root := src
		if pattern == "cmd" {
			root += "cmd" + string(filepath.Separator)
		}
		filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			if err != nil || !fi.IsDir() || path == src {
				return nil
			}

			// Avoid .foo, _foo, and testdata directory trees.
			_, elem := filepath.Split(path)
			if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" {
				return filepath.SkipDir
			}

			name := filepath.ToSlash(path[len(src):])
			if pattern == "std" && (strings.Contains(name, ".") || name == "cmd") {
				// The name "std" is only the standard library.
				// If the name is cmd, it's the root of the command tree.
				return filepath.SkipDir
			}
			if !treeCanMatch(name) {
				return filepath.SkipDir
			}
			if have[name] {
				return nil
			}
			have[name] = true
			if !match(name) {
				return nil
			}
			_, err = buildContext.ImportDir(path, 0)
			if err != nil {
				if _, noGo := err.(*build.NoGoError); noGo {
					return nil
				}
			}
			pkgs = append(pkgs, name)
			return nil
		})
	}
	return pkgs
}

// allPackagesInFS is like allPackages but is passed a pattern
// beginning ./ or ../, meaning it should scan the tree rooted
// at the given directory.  There are ... in the pattern too.
func allPackagesInFS(pattern string) []string {
	pkgs := matchPackagesInFS(pattern)
	if len(pkgs) == 0 {
		// fmt.Fprintf(os.Stderr, "warning: %q matched no packages\n", pattern)
	}
	return pkgs
}

func matchPackagesInFS(pattern string) []string {
	// Find directory to begin the scan.
	// Could be smarter but this one optimization
	// is enough for now, since ... is usually at the
	// end of a path.
	i := strings.Index(pattern, "...")
	dir, _ := pathPkg.Split(pattern[:i])

	// pattern begins with ./ or ../.
	// path.Clean will discard the ./ but not the ../.
	// We need to preserve the ./ for pattern matching
	// and in the returned import paths.
	prefix := ""
	if strings.HasPrefix(pattern, "./") {
		prefix = "./"
	}
	match := matchPattern(pattern)

	var pkgs []string
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}
		if path == dir {
			// filepath.Walk starts at dir and recurses. For the recursive case,
			// the path is the result of filepath.Join, which calls filepath.Clean.
			// The initial case is not Cleaned, though, so we do this explicitly.
			//
			// This converts a path like "./io/" to "io". Without this step, running
			// "cd $GOROOT/src/pkg; go list ./io/..." would incorrectly skip the io
			// package, because prepending the prefix "./" to the unclean path would
			// result in "././io", and match("././io") returns false.
			path = filepath.Clean(path)
		}

		// Avoid .foo, _foo, and testdata directory trees, but do not avoid "." or "..".
		_, elem := filepath.Split(path)
		dot := strings.HasPrefix(elem, ".") && elem != "." && elem != ".."
		if dot || strings.HasPrefix(elem, "_") || elem == "testdata" {
			return filepath.SkipDir
		}

		name := prefix + filepath.ToSlash(path)
		if !match(name) {
			return nil
		}
		if _, err = build.ImportDir(path, 0); err != nil {
			if _, noGo := err.(*build.NoGoError); !noGo {
				// log.Print(err)
			}
			return nil
		}
		pkgs = append(pkgs, name)
		return nil
	})
	return pkgs
}
