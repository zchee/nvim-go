// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/neovim-go/vim"
	"golang.org/x/tools/refactor/rename"
)

const pkgRename = "GoRename"

type cmdRenameEval struct {
	Cwd        string `msgpack:",array"`
	File       string
	RenameFrom string
}

func cmdRename(v *vim.Vim, args []string, bang bool, eval *cmdRenameEval) {
	go Rename(v, args, bang, eval)
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func Rename(v *vim.Vim, args []string, bang bool, eval *cmdRenameEval) error {
	defer profile.Start(time.Now(), "GoRename")

	ctxt := new(context.Context)
	defer ctxt.Build.SetContext(filepath.Dir(eval.File))()

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

	offset, err := nvim.ByteOffset(v, b, w)
	if err != nil {
		return err
	}
	pos := fmt.Sprintf("%s:#%d", eval.File, offset)

	var renameTo string
	if len(args) > 0 {
		renameTo = args[0]
	} else {
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", pkgRename, eval.RenameFrom)
		var toResult interface{}
		if config.RenamePrefill {
			err := v.Call("input", &toResult, askMessage, eval.RenameFrom)
			if err != nil {
				return nvim.EchohlErr(v, pkgRename, "Keyboard interrupt")
			}
		} else {
			err := v.Call("input", &toResult, askMessage)
			if err != nil {
				return nvim.EchohlErr(v, pkgRename, "Keyboard interrupt")
			}
		}
		if toResult.(string) == "" {
			return nvim.EchohlErr(v, pkgRename, "Not enough arguments for rename destination name")
		}
		renameTo = fmt.Sprintf("%s", toResult)
	}

	v.Command(fmt.Sprintf("echo '%s: Renaming ' | echohl Identifier | echon '%s' | echohl None | echon ' to ' | echohl Identifier | echon '%s' | echohl None | echon ' ...'", pkgRename, eval.RenameFrom, renameTo))

	if bang {
		rename.Force = true
	}

	// TODO(zchee): More elegant way
	saveStdout, saveStderr := os.Stdout, os.Stderr
	os.Stderr = os.Stdout
	read, write, _ := os.Pipe()
	os.Stdout, os.Stderr = write, write
	defer func() {
		os.Stderr = saveStdout
		os.Stderr = saveStderr
	}()

	if err := rename.Main(&build.Default, pos, "", renameTo); err != nil {
		write.Close()
		er, _ := ioutil.ReadAll(read)
		log.Printf("er: %+v\n", string(er))
		go func() {
			loclist, _ := quickfix.ParseError(er, eval.Cwd, &ctxt.Build)
			quickfix.SetLoclist(v, loclist)
			quickfix.OpenLoclist(v, w, loclist, true)
		}()

		return nvim.EchohlErr(v, "GoRename", err)
	}

	write.Close()
	out, _ := ioutil.ReadAll(read)
	defer nvim.EchoSuccess(v, pkgRename, fmt.Sprintf("%s", out))

	// TODO(zchee): 'edit' command is ugly.
	// Should create tempfile and use SetBufferLines.
	return v.Command("silent edit")
}
