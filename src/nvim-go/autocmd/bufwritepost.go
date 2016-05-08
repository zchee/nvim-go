package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufWritePost", &plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmdBufWritePost)
}

type bufwritepostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func autocmdBufWritePost(v *vim.Vim, eval bufwritepostEval) error {
	if config.BuildAutosave {
		go commands.Build(v, commands.CmdBuildEval{
			Cwd: eval.Cwd,
			Dir: eval.Dir,
		})
	}

	if config.TestAutosave {
		go commands.Test(v, eval.Dir)
	}

	return nil
}
