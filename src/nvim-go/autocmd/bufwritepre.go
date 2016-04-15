package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
)

func init() {
	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "fmt", Eval: "*"}, autocmdBufWritePre)
}

type fileInfo struct {
	Cwd  string `eval:"getcwd()"`
	Path string `eval:"expand('%:p')"`
	Name string `eval:"expand('%')"`
}

type env struct {
	FmtAsync           int64    `eval:"g:go#fmt#async"`
	IferrAutosave      int64    `eval:"g:go#iferr#autosave"`
	MetaLinterAutosave int64    `eval:"g:go#lint#metalinter#autosave"`
	MetaLinterTools    []string `eval:"g:go#lint#metalinter#autosave#tools"`
	MetaLinterDeadline string   `eval:"g:go#lint#metalinter#deadline"`
}

func autocmdBufWritePre(v *vim.Vim, eval *struct {
	FileInfo fileInfo
	Env      env
}) error {
	if eval.Env.IferrAutosave == int64(1) {
		var env = commands.CmdIferrEval{
			Cwd:  eval.FileInfo.Cwd,
			File: eval.FileInfo.Path,
		}
		go commands.Iferr(v, env)
	}

	if eval.Env.MetaLinterAutosave == int64(1) {
		var env = commands.CmdMetalinterEval{
			Cwd:      eval.FileInfo.Cwd,
			Tools:    eval.Env.MetaLinterTools,
			Deadline: eval.Env.MetaLinterDeadline,
		}
		go commands.Metalinter(v, env)
	}

	if eval.Env.FmtAsync == int64(1) {
		go commands.Fmt(v, eval.FileInfo.Cwd)
	} else {
		return commands.Fmt(v, eval.FileInfo.Cwd)
	}
	return nil
}
