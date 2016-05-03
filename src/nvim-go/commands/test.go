package commands

import (
	"go/build"
	"os"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gotest", &plugin.CommandOptions{Eval: "expand('%:p:h')"}, cmdTest)
}

func cmdTest(v *vim.Vim, dir string) {
	go Test(v, dir)
}

// Test run the package test command use compile tool that determined from
// the directory structure.
func Test(v *vim.Vim, dir string) error {
	defer nvim.Profile(time.Now(), "GoTest")
	var ctxt = context.Build{}
	defer ctxt.SetContext(dir)()

	buildDir := strings.Split(build.Default.GOPATH, ":")[0]
	var cmd []string
	if buildDir == os.Getenv("GOPATH") {
		cmd = append(cmd, "go")
	} else {
		cmd = append(cmd, "gb")
	}
	cmd = append(cmd, "test", "-v", "./...")

	term := nvim.NewTerminal(v, cmd, config.TerminalMode)

	rootDir := context.FindVcsRoot(dir)
	term.Dir = rootDir

	if err := term.Run(); err != nil {
		return err
	}

	return nil
}
