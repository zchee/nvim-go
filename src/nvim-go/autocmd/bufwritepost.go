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

func autocmdBufWritePost(v *vim.Vim, eval commands.CmdBuildEval) error {
	if config.BuildAutosave {
		go commands.Build(v, eval)
	}

	return nil
}
