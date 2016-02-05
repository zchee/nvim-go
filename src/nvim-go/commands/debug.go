package commands

import (
	"encoding/binary"
	"fmt"

	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Godebug", &plugin.CommandOptions{NArgs: "*", Eval: "getcwd()"}, Debug)
}

func Debug(v *vim.Vim, args []string, cwd string) error {
	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	bu, err := v.BufferLineSlice(b, 0, -1, true, true)
	if err != nil {
		return err
	}

	w, _ := v.CurrentWindow()
	cursor, _ := v.WindowCursor(w)
	height, _ := v.WindowHeight(w)
	width, _ := v.WindowWidth(w)
	pos, _ := v.WindowPosition(w)
	win, _ := v.Windows()

	offset := 0
	cursorline := 1
	for _, bytes := range bu {
		if cursor[0] == 1 {
			offset = 1
			break
		} else if cursorline == cursor[0] {
			offset++
			break
		}
		offset += (binary.Size(bytes) + 1)
		cursorline++
	}

	// return command.Echomsg(v, fmt.Sprintf("line: %d col: %d offset: %d", cursor[0], cursor[1], (offset+(cursor[1]-1))))
	return nvim.Echomsg(v, fmt.Sprintf("win: %v height: %v width: %v pos %v", win, height, width, pos))
}
