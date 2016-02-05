package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

func Echomsg(v *vim.Vim, format string, args ...interface{}) error {
	return v.WriteOut(fmt.Sprintf(format, args...))
}

func Echoerror(v *vim.Vim, format string, args ...interface{}) error {
	return v.WriteErr(fmt.Sprintf(format, args...))
}
