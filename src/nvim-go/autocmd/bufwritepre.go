package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
	"nvim-go/config"
)

func init() {
	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmdBufWritePre)
}

type bufwritepreEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func autocmdBufWritePre(v *vim.Vim, eval bufwritepreEval) error {
	if config.IferrAutosave == int64(1) {
		var env = commands.CmdIferrEval{
			Cwd:  eval.Cwd,
			File: eval.File,
		}
		go commands.Iferr(v, env)
	}

	if config.MetalinterAutosave == int64(1) {
		go commands.Metalinter(v, eval.Cwd)
	}

	if config.FmtAsync == int64(1) {
		go commands.Fmt(v, eval.Cwd)
	} else {
		return commands.Fmt(v, eval.Cwd)
	}
	return nil
}
