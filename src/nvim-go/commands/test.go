// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"
	"nvim-go/nvimutil/profile"
	"nvim-go/nvimutil/quickfix"
	"nvim-go/pathutil"

	vim "github.com/neovim/go-client/nvim"
	"golang.org/x/tools/go/ast/astutil"
)

func (c *Commands) cmdTest(args []string, dir string) {
	go c.Test(args, dir)
}

// testTerm cache nvimutil.Terminal use global variable.
var testTerm *nvimutil.Terminal

// Test run the package test command use compile tool that determined from
// the directory structure.
func (c *Commands) Test(args []string, dir string) error {
	defer profile.Start(time.Now(), "GoTest")
	defer c.ctxt.SetContext(dir)()

	cmd := []string{c.ctxt.Build.Tool, "test"}
	args = append(args, config.TestArgs...)
	if len(args) > 0 {
		cmd = append(cmd, args...)
	}

	if c.ctxt.Build.Tool == "go" {
		cmd = append(cmd, string("./..."))
	}

	if testTerm == nil {
		testTerm = nvimutil.NewTerminal(c.v, "__GO_TEST__", cmd, config.TerminalMode)
		testTerm.Dir = pathutil.FindVCSRoot(dir)
	}

	if err := testTerm.Run(cmd); err != nil {
		return err
	}

	return nil
}

var (
	fset       = token.NewFileSet() // *token.FileSet
	parserMode parser.Mode          // uint
	pos        token.Pos

	testPrefix       = "Test"
	testSuffix       = "_test"
	isTest           bool
	funcName         string
	funcNameNoExport string
)

type cmdTestSwitchEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Commands) cmdTestSwitch(eval cmdTestSwitchEval) {
	go c.TestSwitch(eval)
}

// TestSwitch switch to corresponds current cursor (test)function.
func (c *Commands) TestSwitch(eval cmdTestSwitchEval) error {
	// Check the current buffer name whether '*_test.go'.
	fname := eval.File
	exp := filepath.Ext(fname)
	var switchfile string
	if strings.Index(fname, testSuffix) == -1 {
		isTest = false
		switchfile = strings.Replace(fname, exp, testSuffix+exp, 1) // not testfile
	} else {
		isTest = true
		switchfile = strings.Replace(fname, testSuffix+exp, exp, 1) // testfile
	}

	// Check the exists of switch destination file.
	if _, err := os.Stat(switchfile); err != nil {
		return nvimutil.EchohlErr(c.v, "GoTestSwitch", "Switch destination file does not exist")
	}

	dir, _ := filepath.Split(fname)
	defer c.ctxt.SetContext(filepath.Dir(dir))()

	var (
		b vim.Buffer
		w vim.Window
	)

	// Gets the current buffer information.
	if c.p == nil {
		c.p = c.v.NewPipeline()
	}
	c.p.CurrentBuffer(&b)
	c.p.CurrentWindow(&w)
	if err := c.p.Wait(); err != nil {
		return err
	}

	// Get the byte offset of current cursor position from buffer.
	// TODO(zchee): Eval 'line2byte(line('.'))+(col('.')-2)' is faster and safer?
	byteOffset, err := nvimutil.ByteOffset(c.v, b, w)
	if err != nil {
		return err
	}
	// Get the 2d byte slice of current buffer.
	var buf [][]byte
	c.p.BufferLines(b, 0, -1, true, &buf)
	if err := c.p.Wait(); err != nil {
		return err
	}

	f, err := parse(fname, fset, nvimutil.ToByteSlice(buf)) // *ast.File
	if err != nil {
		return err
	}
	offset := fset.File(f.Pos()).Pos(byteOffset) // token.Pos

	// Parses the function ast node from the current cursor position.
	qpos, _ := astutil.PathEnclosingInterval(f, offset, offset) // path []ast.Node, exact bool
	for _, q := range qpos {
		switch x := q.(type) {
		case *ast.FuncDecl:
			if x.Name != nil { // *ast.Ident
				// TODO(zchee): Support parses the function struct name.
				// If the function has a struct, gotests will be generated the
				// mixed camel case test function name include struct name for prefix.
				if !isTest {
					funcName = fmt.Sprintf("%s%s", testPrefix, ToPascalCase(x.Name.Name))
				} else {
					funcName = strings.Replace(x.Name.Name, testPrefix, "", 1)
				}
				funcNameNoExport = ToMixedCase(funcName)
			}
		}
	}

	// Get the switch destination file ast node.
	fswitch, err := parse(switchfile, fset, nil) // *ast.File
	if err != nil {
		return err
	}

	// Reset pos value.
	if pos != token.NoPos {
		pos = 0
	}
	// Parses the switch destination file ast node.
	ast.Walk(visitorFunc(parseFunc), fswitch)

	if !pos.IsValid() {
		return nvimutil.EchohlErr(c.v, "GoTestSwitch", "Not found the switch destination function")
	}

	// Jump to the corresponds function.
	return quickfix.GotoPos(c.v, w, fset.Position(pos), eval.Cwd)
}

// Wrapper of the parser.ParseFile()
func parse(filename string, fset *token.FileSet, src interface{}) (*ast.File, error) {
	file, err := parser.ParseFile(fset, filename, src, parserMode)
	if err != nil {
		return nil, err
	}

	return file, err
}

// visitorFunc for ast.Visit type.
type visitorFunc func(n ast.Node) ast.Visitor

// visit for ast.Visit function.
func (f visitorFunc) Visit(n ast.Node) ast.Visitor {
	return f(n)
}

// Core of the parser of the ast node.
func parseFunc(node ast.Node) ast.Visitor {
	switch x := node.(type) {
	case *ast.FuncDecl:
		if x.Name.Name == funcName || x.Name.Name == funcNameNoExport || indexFuncName(x.Name.Name, funcName, funcNameNoExport) { // x.Name.Name: *ast.Ident.string
			pos = x.Name.NamePos
			return nil
		}
	}

	return visitorFunc(parseFunc)
}

func indexFuncName(s string, sep ...string) bool {
	for _, fn := range sep {
		i := strings.Index(fn, s)
		if i > -1 {
			return true
		}
	}

	return false
}
