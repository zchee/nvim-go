package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
	"nvim-go/vars"
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
	if vars.IferrAutosave == int64(1) {
		var env = commands.CmdIferrEval{
			Cwd:  eval.Cwd,
			File: eval.File,
		}
		go commands.Iferr(v, env)
	}

	if vars.MetalinterAutosave == int64(1) {
		var env = commands.CmdMetalinterEval{
			Cwd:      eval.Cwd,
			Tools:    vars.MetalinterTools,
			Deadline: vars.MetalinterDeadline,
		}
		go commands.Metalinter(v, env)
	}

	if vars.FmtAsync == int64(1) {
		go commands.Fmt(v, eval.Cwd)
	} else {
		return commands.Fmt(v, eval.Cwd)
	}
	return nil
}
