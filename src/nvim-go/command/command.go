package command

import "github.com/garyburd/neovim-go/vim"

func Echomsg(v *vim.Vim, msg string) error {
	return v.Command("echomsg '" + msg + "'")
}
