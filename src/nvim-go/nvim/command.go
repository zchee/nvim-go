package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

func Echomsg(v *vim.Vim, format string, args ...interface{}) error {
	return v.Command("echomsg '" + fmt.Sprintf(format, args...) + "'")
}

func Echoerror(v *vim.Vim, format string, args ...interface{}) error {
	return v.Command("echoerr '" + fmt.Sprintf(format, args...) + "'")
}
