// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/refactor/rename"
)

func init() {
	plugin.HandleCommand("Gorename",
		&plugin.CommandOptions{
			NArgs: "?", Bang: true, Eval: "[getcwd(), expand('%:p:h'), expand('%:p'), expand('<cword>')]"},
		cmdRename)
}

type cmdRenameEval struct {
	Cwd  string `msgpack:",array"`
	Dir  string
	File string
	From string
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

	offset, err := buffer.ByteOffsetPipe(p, b, w)
	if err != nil {
		return err
	}
	pos := fmt.Sprintf("%s:#%d", eval.File, offset)

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
	v.Command(fmt.Sprintf("echon '%s: Renaming ' | echohl Identifier | echon '%s' | echohl None | echon ' to ' | echohl Identifier | echon '%s' | echohl None | echon ' ...'", prefix, eval.From, to))

	if bang {
		rename.Force = true
	}

	os.Stderr = os.Stdout
	saveStdout := os.Stderr
	read, write, _ := os.Pipe()
	os.Stderr = write

	if err := rename.Main(&build.Default, pos, "", to); err != nil {
		write.Close()
		os.Stderr = saveStdout
		er, _ := ioutil.ReadAll(read)
		go func() {
			loclist, _ := quickfix.ParseError(er, eval.Cwd, &ctxt)
			quickfix.SetLoclist(v, loclist)
			quickfix.OpenLoclist(v, w, loclist, true)
		}()

		return nvim.EchohlErr(v, "GoRename", err)
	}

	write.Close()
	os.Stdout = saveStdout
	out, _ := ioutil.ReadAll(read)
	defer nvim.EchoSuccess(v, prefix, fmt.Sprintf("%s", out))

	// TODO(zchee): Create tempfile and use SetBufferLines.
	return v.Command("edit")
}
