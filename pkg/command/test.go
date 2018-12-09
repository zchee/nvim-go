// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// ----------------------------------------------------------------------------
// GoTest

func (c *Command) cmdTest(ctx context.Context, args []string, dir string) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Test(ctx, args, dir)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Test", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		case nil:
			// nothing to do
		}
	}
}

// testTerm cache nvimutil.Terminal use global variable.
var testTerm *nvimutil.Terminal

// Test run the package test command use compile tool that determined from
// the directory structure.
func (c *Command) Test(ctx context.Context, args []string, dir string) error {
	var span *trace.Span
	ctx, span = monitoring.StartSpan(ctx, "Test")
	defer span.End()

	cmd := []string{c.buildContext.Build.Tool, "test", strings.Join(config.TestFlags, " ")}
	if len(args) > 0 {
		cmd = append(cmd, args...)
	}

	var testPkgs []string
	if config.TestAll {
		switch c.buildContext.Build.Tool {
		case "go":
			pkgs, err := fs.FindAllPackage(dir, build.Default, nil, fs.ModeExcludeVendor)
			if err != nil {
				span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
				return errors.WithStack(err)
			}
			for _, p := range pkgs {
				testPkgs = append(testPkgs, fs.TrimGoPath(p.Dir))
			}
		case "gb":
			// nothing to do
		}
	} else {
		pkgs, err := fs.PackageID(dir)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		}
		testPkgs = append(testPkgs, pkgs)
	}

	cmd = append(cmd, testPkgs...)

	if testTerm == nil {
		testTerm = nvimutil.NewTerminal(c.Nvim, "__GO_TEST__", cmd, config.TerminalMode)
		testTerm.Dir = fs.FindVCSRoot(dir)
	}

	if err := testTerm.Run(cmd); err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
	}

	return nil
}

// ----------------------------------------------------------------------------
// GoSwitchTest

var (
	fset       = token.NewFileSet()
	parserMode parser.Mode
	pos        token.Pos

	testPrefix = "Test"
	testSuffix = "_test.go"
	isTest     bool
	funcName   string

	testFileRe = regexp.MustCompile(`Test`)
)

type cmdTestSwitchEval struct {
	Cwd    string `msgpack:",array"`
	File   string
	Offset int
}

// SwitchTest switch to the corresponds current cursor (Test)function.
func (c *Command) SwitchTest(ctx context.Context, eval *cmdTestSwitchEval) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	defer nvimutil.Profile(ctx, time.Now(), "GoSwitchTest")

	fname := eval.File
	ext := filepath.Ext(fname)

	// Checks whether the current buffer name contains '_test.go' and assign
	// destination filename to switchFile
	var switchFile string
	if isTest = strings.Contains(fname, testSuffix); !isTest {
		// code file
		switchFile = strings.Replace(fname, ext, testSuffix, 1)
	} else {
		// test file
		switchFile = strings.Replace(fname, testSuffix, ext, 1)
	}

	// Check exists of switch destination file
	if !fs.IsExist(switchFile) {
		return errors.New("Does not exist the switching destination file")
	}

	b := nvim.Buffer(c.buildContext.BufNr)
	w := nvim.Window(c.buildContext.WinID)

	// Get the 2D byte slice of current buffer
	buf, err := c.Nvim.BufferLines(b, 0, -1, true)
	if err != nil {
		return errors.WithStack(err)
	}

	f := parse(fname, fset, nvimutil.ToByteSlice(buf))
	if f == nil {
		return errors.New("couldn't parse of the current buffer")
	}
	offset := fset.File(f.Pos()).Pos(eval.Offset)

	// Parses the AST node from the current cursor position
	qpos, _ := astutil.PathEnclosingInterval(f, offset, offset)
	for _, q := range qpos {
		switch x := q.(type) {
		case *ast.FuncDecl:
			if x.Name != nil {
				if !isTest {
					funcName = x.Name.Name
				} else {
					funcName = strings.TrimPrefix(x.Name.Name, testPrefix)
				}
				// currentlly only support "_" or "-"
				if funcName[0] == '_' || funcName[0] == '-' {
					funcName = funcName[1:]
				}
			}
		}
	}

	fswitch := parse(switchFile, fset, nil)
	if fswitch == nil {
		return errors.New("couldn't parse of the destination file")
	}

	// Reset pos value
	if pos != token.NoPos {
		pos = 0
	}

	// Find the destination function
	ast.Walk(visitorFunc(matchFunc), fswitch)

	if !pos.IsValid() {
		return nvimutil.Echoerr(c.Nvim, "GoSwitchTest: Not found the switch destination function")
	}

	// Goto the destination file and function position
	return nvimutil.GotoPos(c.Nvim, w, fset.Position(pos), eval.Cwd)
}

// parse wrapper of the parser.ParseFile()
func parse(filename string, fset *token.FileSet, src interface{}) *ast.File {
	file, err := parser.ParseFile(fset, filename, src, parserMode)
	if err != nil {
		return nil
	}

	return file
}

// visitorFunc for ast.Visit type.
type visitorFunc func(n ast.Node) ast.Visitor

// Visit for ast.Visit function.
func (f visitorFunc) Visit(n ast.Node) ast.Visitor {
	return f(n)
}

// matchFunc checks whether the matching funcName in the node.
func matchFunc(node ast.Node) ast.Visitor {
	switch x := node.(type) {
	case *ast.FuncDecl:
		if isTest && x.Recv != nil {
			if recv, ok := x.Recv.List[0].Type.(*ast.StarExpr); ok {
				funcName = strings.TrimPrefix(funcName, recv.X.(*ast.Ident).Name+"_")
			}
		}
		if x.Name.Name == funcName || matchFuncName(x.Name.Name, funcName) {
			pos = x.Name.NamePos
			return nil
		}
	}

	return visitorFunc(matchFunc)
}

// matchFuncName returns whether the matches the function name.
func matchFuncName(s, fn string) bool {
	if ok, err := regexp.MatchString(fmt.Sprintf(`(?i)%s(?:[[:graph:]]*)%s`, testPrefix, fn), s); err == nil && ok {
		return true
	}
	return false
}
