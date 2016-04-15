package commands

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"nvim-go/gb"
	"nvim-go/nvim"
	"os"
	"strings"

	"golang.org/x/tools/refactor/rename"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gorename",
		&plugin.CommandOptions{
			NArgs: "?", Eval: "[expand('%:p:h'), expand('%:p'), line2byte(line('.'))+(col('.')-2)]"},
		cmdRename)
}

var (
	renamePrefill  = "go#rename#prefill"
	vRenamePrefill interface{}
)

type onRenameEval struct {
	Dir    string `msgpack:",array"`
	File   string
	Offset int
}

func cmdRename(v *vim.Vim, args []string, eval *onRenameEval) error {
	go Rename(v, args, eval)
	return nil
}

func Rename(v *vim.Vim, args []string, eval *onRenameEval) error {
	defer gb.WithGoBuildForPath(eval.Dir)()

	v.Var(renamePrefill, &vRenamePrefill)
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
	fmt.Printf(offset)

	askMessage := fmt.Sprintf("%s: Rename '%s' to: ", "nvim-go", from[1:])
	var to interface{}
	if vRenamePrefill.(int64) == int64(1) {
		p.Call("input", &to, askMessage, from[1:])
		if err := p.Wait(); err != nil {
			return nvim.Echomsg(v, "%s", err)
		}
	} else {
		p.Call("input", &to, askMessage)
		if err := p.Wait(); err != nil {
			return nvim.Echomsg(v, "%s", err)
		}
	}

	// out, err := ParseStdout(fmt.Sprintln(eval.Offset))
	// if err != nil {
	// 	return nvim.Echomsg(v, "%s", err)
	// }

	if err := rename.Main(&build.Default, offset, "", fmt.Sprint(to)); err != nil {
		if err != rename.ConflictError {
			nvim.Echomsg(v, "%s", err)
		}
	}
	p.Command("edit!")
	// nvim.Echohl(v, nvim.PackageName, "Function", out)

	return p.Wait()
}

func ParseStdout(trim string) (string, error) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	outC := make(chan string)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	r.Close()
	if err != nil {
		return "", err
	}
	outC <- buf.String()

	w.Close()
	os.Stdout = stdout
	out := <-outC

	ntext := strings.SplitN(out, "#", 2)
	return strings.Trim(ntext[1], trim), nil
}
