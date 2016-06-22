// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package analyze

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"time"

	"nvim-go/commands"
	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

const pkgAnalyzeView = "GoAnalyzeView"

type cmdAnalyzeViewEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func cmdAnalyzeView(v *vim.Vim, eval *cmdAnalyzeViewEval) {
	go analyzeView(v, eval)
}

type analyzeViewer struct {
	data []byte
}

// viewBuffer global variable with cache buffer.
var viewBuffer *buffer.Buffer

// analyzeView gets the Go AST informations of current buffer.
func analyzeView(v *vim.Vim, eval *cmdAnalyzeViewEval) error {
	defer profile.Start(time.Now(), pkgAnalyzeView)

	dir, _ := filepath.Split(eval.File)
	ctxt := new(context.Build)
	defer ctxt.SetContext(dir)()

	var (
		b vim.Buffer
		w vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAnalyzeView)
		return nvim.ErrorWrap(v, err)
	}

	bufferLines, err := v.BufferLineCount(b)
	if err != nil {
		err = errors.Annotate(err, pkgAnalyzeView)
		return nvim.ErrorWrap(v, err)
	}

	src := make([][]byte, bufferLines)
	p.BufferLines(b, 0, -1, true, &src)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAnalyzeView)
		return nvim.ErrorWrap(v, err)
	}

	buf := bytes.Join(src, []byte{'\n'})
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, eval.File, buf, parser.AllErrors|parser.ParseComments)
	if err != nil {
		err = errors.Annotate(err, pkgAnalyzeView)
		return nvim.ErrorWrap(v, err)
	}

	viewer := &analyzeViewer{data: commands.Stringtoslicebyte(fmt.Sprintf("%s Files: %v\n", config.AnalyzeFoldIcon, filepath.Base(eval.File)))}

	ast.Walk(VisitorFunc(viewer.parseBuffer), f)

	data := buffer.ToBufferLines(v, bytes.TrimSuffix(viewer.data, []byte{'\n'}))

	if viewBuffer == nil {
		bufOption := viewOption("buffer")
		bufVar := viewVar("buffer")
		winOption := viewOption("window")
		viewBuffer = buffer.NewBuffer("__GoASTView__", fmt.Sprintf("silent belowright %d vsplit", config.TerminalWidth), int(config.TerminalWidth))
		viewBuffer.Create(v, bufOption, bufVar, winOption, nil)
		viewBuffer.UpdateSyntax(v, "goastview")

		nnoremap := make(map[string]string)
		nnoremap["q"] = ":<C-u>quit<CR>"
		viewBuffer.SetMapping(v, buffer.NoremapNormal, nnoremap)
		p.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")
	}

	v.SetBufferOption(viewBuffer.Buffer, "modifiable", true)
	defer v.SetBufferOption(viewBuffer.Buffer, "modifiable", false)

	p.SetBufferLines(viewBuffer.Buffer, 0, -1, true, data)
	if err := p.Wait(); err != nil {
		err = errors.Annotate(err, pkgAnalyzeView)
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

func (a *analyzeViewer) parseBuffer(node ast.Node) ast.Visitor {
	switch node := node.(type) {

	default:
		return VisitorFunc(a.parseBuffer)
	case *ast.Ident:
		info := fmt.Sprintf("%s *ast.Ident\n\t Name: %v\n\t NamePos: %v\n", config.AnalyzeFoldIcon, node.Name, node.NamePos)
		if fmt.Sprint(node.Obj) != "<nil>" {
			info += fmt.Sprintf("\t Obj: %v\n", node.Obj)
		}
		a.data = append(a.data, commands.Stringtoslicebyte(info)...)
		return VisitorFunc(a.parseBuffer)
	case *ast.GenDecl:
		a.data = append(a.data,
			commands.Stringtoslicebyte(fmt.Sprintf("%s Decls: []ast.Decl\n\t TokPos: %v\n\t Tok: %v\n\t Lparen: %v\n",
				config.AnalyzeFoldIcon, node.TokPos, node.Tok, node.Lparen))...)
		return VisitorFunc(a.parseBuffer)
	case *ast.BasicLit:
		a.data = append(a.data,
			commands.Stringtoslicebyte(fmt.Sprintf("\t- Path: *ast.BasicLit\n\t\t\t Value: %v\n\t\t\t Kind: %v\n\t\t\t ValuePos: %v\n",
				node.Value, node.Kind, node.ValuePos))...)
		return VisitorFunc(a.parseBuffer)

	}
}

func viewOption(scope string) map[string]interface{} {
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

func viewVar(scope string) map[string]interface{} {
	vars := make(map[string]interface{})

	switch scope {
	case "buffer":
		vars[buffer.Colorcolumn] = ""
	}

	return vars
}
