package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
	"nvim-go/vars"
)

func init() {
	plugin.HandleAutocmd("BufWritePost", &plugin.AutocmdOptions{Pattern: "*.go", Eval: "expand('%:p:h')"}, autocmdBufWritePost)
}

func autocmdBufWritePost(v *vim.Vim, cwd string) error {
	if vars.BuildAutosave != int64(0) {
		go commands.Build(v, cwd)
	}

	return nil
}
