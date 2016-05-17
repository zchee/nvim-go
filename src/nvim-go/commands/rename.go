package commands

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/refactor/rename"
)

func init() {
	plugin.HandleCommand("Gorename",
		&plugin.CommandOptions{
			NArgs: "?", Bang: true, Eval: "[expand('%:p:h'), expand('%:p'), line2byte(line('.'))+(col('.')-2), expand('<cword>')]"},
		cmdRename)
}

type cmdRenameEval struct {
	Dir    string `msgpack:",array"`
	File   string
	Offset int
	From   string
}

func cmdRename(v *vim.Vim, args []string, bang bool, eval *cmdRenameEval) {
	go Rename(v, args, bang, eval)
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func Rename(v *vim.Vim, args []string, bang bool, eval *cmdRenameEval) error {
	defer profile.Start(time.Now(), "GoRename")
	var ctxt = context.Build{}
	defer ctxt.SetContext(eval.Dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	offset := fmt.Sprintf("%s:#%d", eval.File, eval.Offset)

	var to string
	if len(args) > 0 {
		to = args[0]
	} else {
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", "GoRename", eval.From)
		var toResult interface{}
		if config.RenamePrefill {
			err := v.Call("input", &toResult, askMessage, eval.From)
			if err != nil {
				return nvim.EchohlErr(v, "GoRename", "Keyboard interrupt")
			}
		} else {
			err := v.Call("input", &toResult, askMessage)
			if err != nil {
				return nvim.EchohlErr(v, "GoRename", "Keyboard interrupt")
			}
		}
		if toResult.(string) == "" {
			return nvim.EchohlErr(v, "GoRename", "Not enough arguments for rename destination name")
		}
		to = fmt.Sprintf("%s", toResult)
	}

	prefix := "GoRename"
	nvim.EchoProgress(v, prefix, "Renaming", eval.From, to)

	if bang {
		rename.Force = true
	}

	saveStdout := os.Stdout
	saveStderr := os.Stderr
	ro, wo, _ := os.Pipe()
	re, we, _ := os.Pipe()
	os.Stdout = wo
	os.Stderr = we

	if err := rename.Main(&build.Default, offset, "", to); err != nil {
		return err
	}

	wo.Close()
	we.Close()
	os.Stdout = saveStdout
	os.Stderr = saveStderr

	out, _ := ioutil.ReadAll(ro)
	er, _ := ioutil.ReadAll(re)
	if len(er) != 0 {
		nvim.EchohlErr(v, "GoRename", er)
	}
	defer nvim.EchoSuccess(v, prefix, fmt.Sprintf("%s", out))

	// TODO(zchee): Create tempfile and use SetBufferLines.
	return v.Command("edit")
}
