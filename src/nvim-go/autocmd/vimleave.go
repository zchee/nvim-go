package autocmd

import (
	"nvim-go/commands/delve"

	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("VimLeavePre", &plugin.AutocmdOptions{Group: "nvim-go", Pattern: "*"}, delve.CmdDelveDetach)
}
