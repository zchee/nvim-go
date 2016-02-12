package commands

import (
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Godebug", &plugin.CommandOptions{NArgs: "*", Eval: "getcwd()"}, Debug)
}

func Debug(v *vim.Vim, args []string, cwd string) error {
	w, _ := v.CurrentWindow()
	// b, _ := v.CurrentBuffer()
	// l, _ := v.CurrentLine()
	// t, _ := v.CurrentTabpage()

	// var vars interface{}
	// v.Var("go#test#vars", &vars)
	// fmt.Println(vars)
	// return nvim.Echomsg(v, "loaded")

	win, _ := v.Windows()
	height, _ := v.WindowHeight(w)
	width, _ := v.WindowWidth(w)
	pos, _ := v.WindowPosition(w)
	return nvim.Echomsg(v, "win: %v height: %v width: %v pos %v", win, height, width, pos)
}
