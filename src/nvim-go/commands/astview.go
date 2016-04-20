package commands

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"unsafe"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

var ASTInfo []byte

func init() {
	plugin.HandleCommand("GoAstView", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p')]"}, funcAstView)
}

type funcAstEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func funcAstView(v *vim.Vim, eval *funcAstEval) {
	go AstView(v, eval)
}

func AstView(v *vim.Vim, eval *funcAstEval) error {
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

	var sources [][]byte
	p.BufferLines(b, 0, -1, false, &sources)
	if err := p.Wait(); err != nil {
		return err
	}

	var buf []byte
	for _, b := range sources {
		buf = append(buf, b...)
		buf = append(buf, byte('\n'))
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, eval.File, buf, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	_, file := filepath.Split(eval.File)
	ASTInfo = append(ASTInfo, stringtoslicebyte(fmt.Sprintf("▼ Files: %v\n", file))...)
	ast.Walk(VisitorFunc(parseAST), f)

	p.Command("vertical botright 80 new")
	p.CurrentBuffer(&b)
	if err := p.Wait(); err != nil {
		return err
	}

	out := bytes.Split(bytes.TrimSuffix(ASTInfo, []byte{'\n'}), []byte{'\n'})
	p.SetBufferLines(b, 0, -1, true, out)

	p.SetBufferName(b, "__GoAstView__")
	p.SetBufferOption(b, "buftype", "nofile")
	p.SetBufferOption(b, "filetype", "goastview")
	p.SetBufferOption(b, "nonumber", true)
	p.SetBufferOption(b, "nolist", true)
	p.SetBufferOption(b, "colorcolumn", "")
	p.Command("runtime! syntax/goastview.vim")
	p.Command("set nonumber nolist")
	if err := p.Wait(); err != nil {
		return err
	}

	return p.Wait()
}

type VisitorFunc func(n ast.Node) ast.Visitor

func (f VisitorFunc) Visit(n ast.Node) ast.Visitor {
	return f(n)
}

func parseAST(node ast.Node) ast.Visitor {
	switch node := node.(type) {

	default:
		return VisitorFunc(parseAST)
	case *ast.Ident:
		ASTInfo = append(ASTInfo,
			stringtoslicebyte(fmt.Sprintf("▼ *ast.Ident\n\t Name: %v\n\t NamePos: %v\n\t Obj: %v\n",
				node.Name, node.NamePos, node.Obj))...)
		return VisitorFunc(parseAST)
	case *ast.GenDecl:
		ASTInfo = append(ASTInfo,
			stringtoslicebyte(fmt.Sprintf("▼ Decls: []ast.Decl\n\t TokPos: %v\n\t Tok: %v\n\t Lparen: %v\n",
				node.TokPos, node.Tok, node.Lparen))...)
		return VisitorFunc(parseAST)
	case *ast.BasicLit:
		ASTInfo = append(ASTInfo,
			stringtoslicebyte(fmt.Sprintf("\t- Path: *ast.BasicLit\n\t\t\t Value: %v\n\t\t\t Kind: %v\n\t\t\t ValuePos: %v\n",
				node.Value, node.Kind, node.ValuePos))...)
		return VisitorFunc(parseAST)

	}

	return nil
}

func stringtoslicebyte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
