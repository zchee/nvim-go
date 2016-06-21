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

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
)

const pkgAstView = "GoASTView"

func init() {
	plugin.HandleCommand("GoAstView", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p')]"}, cmdASTView)
}

type cmdAstEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func cmdASTView(v *vim.Vim, eval *cmdAstEval) {
	go astView(v, eval)
}

type astViewer struct {
	data []byte
}

// astBuffer global variable with cache buffer.
var astViewBuffer *buffer.Buffer

// AstView gets the Go AST informations of current buffer.
func astView(v *vim.Vim, eval *cmdAstEval) error {
	defer profile.Start(time.Now(), "AstView")

	var (
		b vim.Buffer
		w vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAstView)
		return nvim.ErrorWrap(v, err)
	}

	bufferLines, err := v.BufferLineCount(b)
	if err != nil {
		err = errors.Annotate(err, pkgAstView)
		return nvim.ErrorWrap(v, err)
	}

	src := make([][]byte, bufferLines)
	p.BufferLines(b, 0, -1, true, &src)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAstView)
		return nvim.ErrorWrap(v, err)
	}

	buf := bytes.Join(src, []byte{'\n'})
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, eval.File, buf, parser.AllErrors|parser.ParseComments)
	if err != nil {
		err = errors.Annotate(err, pkgAstView)
		return nvim.ErrorWrap(v, err)
	}

	_, file := filepath.Split(eval.File)
	viewer := &astViewer{data: stringtoslicebyte(fmt.Sprintf("%s Files: %v\n", config.AstFoldIcon, file))}

	ast.Walk(VisitorFunc(viewer.parseAST), f)

	data := buffer.ToBufferLines(v, bytes.TrimSuffix(viewer.data, []byte{'\n'}))

	if astViewBuffer == nil {
		bufOption := astViewOption("buffer")
		bufVar := astViewVar("buffer")
		winOption := astViewOption("window")
		astViewBuffer = buffer.NewBuffer("__GoASTView__", fmt.Sprintf("silent belowright %d vsplit", config.TerminalWidth), int(config.TerminalWidth))
		astViewBuffer.Create(v, bufOption, bufVar, winOption, nil)
		astViewBuffer.UpdateSyntax(v, "goastview")

		nnoremap := make(map[string]string)
		nnoremap["q"] = ":<C-u>quit<CR>"
		astViewBuffer.SetMapping(v, buffer.NoremapNormal, nnoremap)
		p.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")
	}

	v.SetBufferOption(astViewBuffer.Buffer, "modifiable", true)
	defer v.SetBufferOption(astViewBuffer.Buffer, "modifiable", false)

	p.SetBufferLines(astViewBuffer.Buffer, 0, -1, true, data)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAstView)
		return nvim.ErrorWrap(v, err)
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

func (a *astViewer) parseAST(node ast.Node) ast.Visitor {
	switch node := node.(type) {

	default:
		return VisitorFunc(a.parseAST)
	case *ast.Ident:
		info := fmt.Sprintf("%s *ast.Ident\n\t Name: %v\n\t NamePos: %v\n", config.AstFoldIcon, node.Name, node.NamePos)
		if fmt.Sprint(node.Obj) != "<nil>" {
			info += fmt.Sprintf("\t Obj: %v\n", node.Obj)
		}
		a.data = append(a.data, stringtoslicebyte(info)...)
		return VisitorFunc(a.parseAST)
	case *ast.GenDecl:
		a.data = append(a.data,
			stringtoslicebyte(fmt.Sprintf("%s Decls: []ast.Decl\n\t TokPos: %v\n\t Tok: %v\n\t Lparen: %v\n",
				config.AstFoldIcon, node.TokPos, node.Tok, node.Lparen))...)
		return VisitorFunc(a.parseAST)
	case *ast.BasicLit:
		a.data = append(a.data,
			stringtoslicebyte(fmt.Sprintf("\t- Path: *ast.BasicLit\n\t\t\t Value: %v\n\t\t\t Kind: %v\n\t\t\t ValuePos: %v\n",
				node.Value, node.Kind, node.ValuePos))...)
		return VisitorFunc(a.parseAST)

	}
}

func astViewOption(scope string) map[string]interface{} {
	options := make(map[string]interface{})

	switch scope {
	case "buffer":
		options[buffer.Bufhidden] = buffer.BufhiddenDelete
		options[buffer.Buflisted] = false
		options[buffer.Buftype] = buffer.BuftypeNofile
		options[buffer.Filetype] = buffer.FiletypeAST
		options[buffer.OpModifiable] = false
		options[buffer.Swapfile] = false
		options[buffer.OpModifiable] = false
	case "window":
		options[buffer.List] = false
		options[buffer.Number] = false
		options[buffer.Relativenumber] = false
		options[buffer.Winfixheight] = false
	}

	return options
}

func astViewVar(scope string) map[string]interface{} {
	vars := make(map[string]interface{})

	switch scope {
	case "buffer":
		vars[buffer.Colorcolumn] = ""
	}

	return vars
}
