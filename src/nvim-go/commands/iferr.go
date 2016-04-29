package commands

import (
	"bufio"
	"go/parser"
	"go/printer"
	"os"

	"nvim-go/context"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/motemen/go-iferr"
	"golang.org/x/tools/go/loader"
)

func init() {
	plugin.HandleCommand("GoIferr", &plugin.CommandOptions{Eval: "[expand('%:p:h'), expand('%:p')]"}, Iferr)
}

// CmdIferrEval type struct for func Iferr().
type CmdIferrEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

// Iferr automatically insert 'if err' Go idiom by parse the current buffer's Go abstract syntax tree(AST).
func Iferr(v *vim.Vim, eval CmdIferrEval) error {
	defer context.WithGoBuildForPath(eval.Cwd)()

	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	bufline, err := v.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}

	var buf string
	for _, bufstr := range bufline {
		buf += "\n" + string(bufstr)
	}

	conf := loader.Config{
		AllowErrors: true,
		ParserMode:  parser.ParseComments,
	}

	f, err := conf.ParseFile(eval.File, buf)
	if err != nil {
		return nvim.Echoerr(v, "GoIferr: %v", err)
	}

	conf.CreateFromFiles(eval.File, f)
	prog, err := conf.Load()
	if err != nil {
		return err
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, pkg := range prog.InitialPackages() {
		for _, f := range pkg.Files {
			iferr.RewriteFile(prog.Fset, f, pkg.Info)
			printer.Fprint(w, prog.Fset, f)
		}
	}

	w.Close()
	os.Stdout = oldStdout

	var out [][]byte
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		out = append(out, scan.Bytes())
	}

	p := v.NewPipeline()
	p.SetBufferLines(b, 0, -1, false, out)

	return p.Wait()
}
