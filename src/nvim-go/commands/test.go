// Copyright 2016 Koichi Shiraishi. All rights reserved.

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
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"
	"nvim-go/nvim/terminal"
	"nvim-go/pathutil"

	"github.com/neovim-go/vim"
	"golang.org/x/tools/go/ast/astutil"
)

func cmdTest(v *vim.Vim, args []string, dir string) {
	go Test(v, args, dir)
}

var term *terminal.Terminal

// Test run the package test command use compile tool that determined from
// the directory structure.
func Test(v *vim.Vim, args []string, dir string) error {
	defer profile.Start(time.Now(), "GoTest")
	ctxt := new(context.Context)
	defer ctxt.Build.SetContext(dir)()

	cmd := []string{ctxt.Build.Tool, "test"}
	args = append(args, config.TestArgs...)
	if len(args) > 0 {
		cmd = append(cmd, args...)
	}

	if ctxt.Build.Tool == "go" {
		cmd = append(cmd, string("./..."))
	}

	if term == nil {
		term = terminal.NewTerminal(v, "__GO_TEST__", cmd, config.TerminalMode)
		term.Dir = pathutil.FindVcsRoot(dir)
	}

	if err := term.Run(cmd); err != nil {
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

func cmdTestSwitch(v *vim.Vim, eval cmdTestSwitchEval) {
	go TestSwitch(v, eval)
}

// TestSwitch switch to corresponds current cursor (test)function.
func TestSwitch(v *vim.Vim, eval cmdTestSwitchEval) error {
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
		return nvim.EchohlErr(v, "GoTestSwitch", "Switch destination file does not exist")
	}

	ctxt := new(context.Context)
	dir, _ := filepath.Split(fname)
	defer ctxt.Build.SetContext(filepath.Dir(dir))()

	var (
		b vim.Buffer
		w vim.Window
	)

	// Gets the current buffer information.
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	// Get the byte offset of current cursor position from buffer.
	// TODO(zchee): Eval 'line2byte(line('.'))+(col('.')-2)' is faster and safer?
	byteOffset, err := nvim.ByteOffset(v, b, w)
	if err != nil {
		return err
	}
	// Get the 2d byte slice of current buffer.
	var buf [][]byte
	p.BufferLines(b, 0, -1, true, &buf)
	if err := p.Wait(); err != nil {
		return err
	}

	f, err := parse(fname, fset, nvim.ToByteSlice(buf)) // *ast.File
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
		return nvim.EchohlErr(v, "GoTestSwitch", "Not found the switch destination function")
	}

	// Jump to the corresponds function.
	return quickfix.GotoPos(v, w, fset.Position(pos), eval.Cwd)
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
