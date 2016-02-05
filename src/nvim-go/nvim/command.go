package nvim

import "github.com/garyburd/neovim-go/vim"

func Echomsg(v *vim.Vim, msg string) error {
	return v.Command("echomsg '" + msg + "'")
}

func Echoerror(v *vim.Vim, msg string) error {
	return v.Command("echoerr '" + msg + "'")
}
