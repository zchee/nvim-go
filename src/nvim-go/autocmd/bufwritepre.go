package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h'), expand('%:p')]"}, autocmdBufWritePre)
}

type bufwritepreEval struct {
	Cwd  string `msgpack:",array"`
	Dir  string
	File string
}

func autocmdBufWritePre(v *vim.Vim, eval bufwritepreEval) error {
	if config.IferrAutosave {
		var env = commands.CmdIferrEval{
			Cwd:  eval.Cwd,
			File: eval.File,
		}
		go commands.Iferr(v, env)
	}

	if config.MetalinterAutosave {
		go commands.Metalinter(v, eval.Cwd)
	}

	if config.FmtAsync {
		go commands.Fmt(v, eval.Dir)
	} else {
		return commands.Fmt(v, eval.Dir)
	}

	return nil
}
