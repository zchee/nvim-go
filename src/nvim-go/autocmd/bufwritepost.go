package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
	"nvim-go/config"
)

func init() {
	plugin.HandleAutocmd("BufWritePost", &plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "expand('%:p:h')"}, autocmdBufWritePost)
}

func autocmdBufWritePost(v *vim.Vim, cwd string) error {
	if config.BuildAutosave {
		go commands.Build(v, cwd)
	}

	return nil
}
