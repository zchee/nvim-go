package commands

import (
	"fmt"
	"go/build"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/refactor/rename"
)

func init() {
	plugin.HandleCommand("Gorename",
		&plugin.CommandOptions{
			NArgs: "?", Eval: "[expand('%:p:h'), expand('%:p'), line2byte(line('.'))+(col('.')-2)]"},
		cmdRename)
}

type cmdRenameEval struct {
	Dir    string `msgpack:",array"`
	File   string
	Offset int
}

func cmdRename(v *vim.Vim, args []string, eval *cmdRenameEval) {
	go Rename(v, args, eval)
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func Rename(v *vim.Vim, args []string, eval *cmdRenameEval) error {
	defer nvim.Profile(time.Now(), "GoRename")
	var ctxt = context.Build{}
	defer ctxt.SetContext(eval.Dir)()

	from, err := v.CommandOutput(fmt.Sprintf("silent! echo expand('<cword>')"))
	if err != nil {
		nvim.Echomsg(v, "%s", err)
	}

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
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", "nvim-go", from[1:])
		var toResult interface{}
		if config.RenamePrefill {
			p.Call("input", &toResult, askMessage, from[1:])
			if err := p.Wait(); err != nil {
				return nvim.Echomsg(v, "%s", err)
			}
		} else {
			p.Call("input", &toResult, askMessage)
			if err := p.Wait(); err != nil {
				return nvim.Echomsg(v, "%s", err)
			}
		}
		if toResult.(string) != "" {
			to = toResult.(string)
		}
	}

	nvim.Echohl(v, "nvim-go: ", "Identifier", "renaming ...")
	if err := rename.Main(&build.Default, offset, "", to); err != nil {
		if err != rename.ConflictError {
			nvim.Echomsg(v, "%s", err)
		}
	}
	defer nvim.ClearMsg(v)
	p.Command("silent! edit!")

	return p.Wait()
}
