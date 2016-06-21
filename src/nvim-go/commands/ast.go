// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"time"
	"unsafe"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
)

const pkgAstView = "AstView"

var (
	astInfo []byte
)

func init() {
	plugin.HandleCommand("GoAstView", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p')]"}, cmdAstView)
}

type cmdAstEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func cmdAstView(v *vim.Vim, eval *cmdAstEval) {
	go astView(v, eval)
}

// AstView gets the Go AST informations of current buffer.
func astView(v *vim.Vim, eval *cmdAstEval) error {
	defer profile.Start(time.Now(), "AstView")

	var (
		b   vim.Buffer
		w   vim.Window
		blc int
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	err := p.Wait()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	sources := make([][]byte, blc)
	p.BufferLines(b, 0, -1, true, &sources)
	p.BufferLineCount(b, &blc)
	if err := p.Wait(); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	var buf []byte
	for _, b := range sources {
		buf = append(buf, b...)
		buf = append(buf, byte('\n'))
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, eval.File, buf, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	_, file := filepath.Split(eval.File)
	astInfo = append(astInfo, stringtoslicebyte(fmt.Sprintf("%s Files: %v\n", config.AstFoldIcon, file))...)
	ast.Walk(VisitorFunc(parseAST), f)

	astinfo := bytes.Split(bytes.TrimSuffix(astInfo, []byte{'\n'}), []byte{'\n'})
	if err := p.Wait(); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	var (
		astBuf vim.Buffer
		astWin vim.Window
	)

	p.Command("vertical botright 80 new")
	p.CurrentBuffer(&astBuf)
	p.CurrentWindow(&astWin)
	if err := p.Wait(); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	p.SetWindowOption(astWin, "number", false)
	p.SetWindowOption(astWin, "list", false)
	p.SetWindowOption(astWin, "colorcolumn", "")

	p.SetBufferName(astBuf, "__GoAstView__")
	p.SetBufferOption(astBuf, "modifiable", true)
	p.SetBufferLines(astBuf, 0, -1, true, astinfo)
	p.SetBufferOption(astBuf, "buftype", "nofile")
	p.SetBufferOption(astBuf, "bufhidden", "delete")
	p.SetBufferOption(astBuf, "buflisted", false)
	p.SetBufferOption(astBuf, "swapfile", false)
	p.SetBufferOption(astBuf, "modifiable", false)
	p.SetBufferOption(astBuf, "filetype", "goastview")
	p.Command("runtime! syntax/goastview.vim")
	if err := p.Wait(); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgAstView))
	}

	p.SetCurrentWindow(w)

	return p.Wait()
}

// VisitorFunc for ast.Visit type.
type VisitorFunc func(n ast.Node) ast.Visitor

// Visit for ast.Visit function.
func (f VisitorFunc) Visit(n ast.Node) ast.Visitor {
	return f(n)
}

func parseAST(node ast.Node) ast.Visitor {
	switch node := node.(type) {

	default:
		return VisitorFunc(parseAST)
	case *ast.Ident:
		info := fmt.Sprintf("%s *ast.Ident\n\t Name: %v\n\t NamePos: %v\n", config.AstFoldIcon, node.Name, node.NamePos)
		if fmt.Sprint(node.Obj) != "<nil>" {
			info += fmt.Sprintf("\t Obj: %v\n", node.Obj)
		}
		astInfo = append(astInfo, stringtoslicebyte(info)...)
		return VisitorFunc(parseAST)
	case *ast.GenDecl:
		astInfo = append(astInfo,
			stringtoslicebyte(fmt.Sprintf("%s Decls: []ast.Decl\n\t TokPos: %v\n\t Tok: %v\n\t Lparen: %v\n",
				config.AstFoldIcon, node.TokPos, node.Tok, node.Lparen))...)
		return VisitorFunc(parseAST)
	case *ast.BasicLit:
		astInfo = append(astInfo,
			stringtoslicebyte(fmt.Sprintf("\t- Path: *ast.BasicLit\n\t\t\t Value: %v\n\t\t\t Kind: %v\n\t\t\t ValuePos: %v\n",
				node.Value, node.Kind, node.ValuePos))...)
		return VisitorFunc(parseAST)

	}
}

func stringtoslicebyte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
