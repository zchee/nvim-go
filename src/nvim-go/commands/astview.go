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

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

var (
	ASTInfo []byte
)

func init() {
	plugin.HandleCommand("GoAstView", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p')]"}, cmdAstView)
}

type cmdAstEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func cmdAstView(v *vim.Vim, eval *cmdAstEval) {
	go AstView(v, eval)
}

func AstView(v *vim.Vim, eval *cmdAstEval) error {
	defer nvim.Profile(time.Now(), "AstView")

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
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
	ASTInfo = append(ASTInfo, stringtoslicebyte(fmt.Sprintf("%s Files: %v\n", config.AstFoldIcon, file))...)
	ast.Walk(VisitorFunc(parseAST), f)

	astinfo := bytes.Split(bytes.TrimSuffix(ASTInfo, []byte{'\n'}), []byte{'\n'})
	if err := p.Wait(); err != nil {
		return err
	}

	p.Command("vertical botright 80 new")
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	p.SetWindowOption(w, "number", false)
	p.SetWindowOption(w, "list", false)
	p.SetWindowOption(w, "colorcolumn", "")

	p.SetBufferName(b, "__GoAstView__")
	p.SetBufferOption(b, "modifiable", true)
	p.SetBufferLines(b, 0, -1, true, astinfo)
	p.SetBufferOption(b, "buftype", "nofile")
	p.SetBufferOption(b, "bufhidden", "delete")
	p.SetBufferOption(b, "buflisted", false)
	p.SetBufferOption(b, "swapfile", false)
	p.SetBufferOption(b, "modifiable", false)
	p.SetBufferOption(b, "filetype", "goastview")
	p.Command("runtime! syntax/goastview.vim")

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
		info := fmt.Sprintf("%s *ast.Ident\n\t Name: %v\n\t NamePos: %v\n", config.AstFoldIcon, node.Name, node.NamePos)
		if fmt.Sprint(node.Obj) != "<nil>" {
			info += fmt.Sprintf("\t Obj: %v\n", node.Obj)
		}
		ASTInfo = append(ASTInfo, stringtoslicebyte(info)...)
		return VisitorFunc(parseAST)
	case *ast.GenDecl:
		ASTInfo = append(ASTInfo,
			stringtoslicebyte(fmt.Sprintf("%s Decls: []ast.Decl\n\t TokPos: %v\n\t Tok: %v\n\t Lparen: %v\n",
				config.AstFoldIcon, node.TokPos, node.Tok, node.Lparen))...)
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
