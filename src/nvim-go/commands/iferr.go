package commands

import (
	"go/format"
	"go/parser"
	"os"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/motemen/go-iferr"
	"golang.org/x/tools/go/loader"

	"nvim-go/gb"
	"nvim-go/nvim"
)

var (
	iferrAsync  = "go#iferr#async"
	vIferrAsync interface{}
)

func init() {
	plugin.HandleCommand("GoIferr", &plugin.CommandOptions{Eval: "[expand('%:p:h'), expand('%:p')]"}, Iferr)
}

type onIferrEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func Iferr(v *vim.Vim, eval onIferrEval) error {
	defer gb.WithGoBuildForPath(eval.Cwd)()

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
		nvim.Echoerr(v, err)
	}

	conf.CreateFromFiles(eval.File, f)
	prog, err := conf.Load()
	if err != nil {
		return err
	}

	// if err != nil {
	// 	nvim.Echoerr(v, err)
	// }
	// fmt.Println(prog.Fset)

	for _, pkg := range prog.InitialPackages() {
		for _, f := range pkg.Files {
			filename := prog.Fset.File(f.Pos()).Name()
			iferr.RewriteFile(prog.Fset, f, pkg.Info)
			fh, err := os.Create(filename)
			if err != nil {
				return err
			}

			format.Node(fh, prog.Fset, f)
		}
	}
	return v.Command("edit!")
}
