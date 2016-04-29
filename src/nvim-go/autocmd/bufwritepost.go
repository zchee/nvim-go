package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufWritePost", &plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "getcwd()"}, autocmdBufWritePost)
}

func autocmdBufWritePost(v *vim.Vim, cwd string) error {
	if config.BuildAutosave {
		go commands.Build(v, cwd)
	}

	return nil
}
