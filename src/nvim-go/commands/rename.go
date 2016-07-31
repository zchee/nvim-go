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
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/neovim-go/vim"
	"github.com/pkg/errors"
	"golang.org/x/tools/refactor/rename"
)

const pkgRename = "GoRename"

type cmdRenameEval struct {
	Cwd        string `msgpack:",array"`
	File       string
	RenameFrom string
}

func (c *Commands) cmdRename(args []string, bang bool, eval *cmdRenameEval) {
	go c.Rename(args, bang, eval)
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func (c *Commands) Rename(args []string, bang bool, eval *cmdRenameEval) error {
	defer profile.Start(time.Now(), "GoRename")
	dir := filepath.Dir(eval.File)
	defer c.ctxt.SetContext(dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	if c.p == nil {
		c.p = c.v.NewPipeline()
	}
	c.p.CurrentBuffer(&b)
	c.p.CurrentWindow(&w)
	if err := c.p.Wait(); err != nil {
		err = errors.Wrap(err, pkgRename)
		return nvim.ErrorWrap(c.v, err)
	}

	offset, err := nvim.ByteOffset(c.v, b, w)
	if err != nil {
		err = errors.Wrap(err, pkgRename)
		return nvim.ErrorWrap(c.v, err)
	}
	pos := fmt.Sprintf("%s:#%d", eval.File, offset)

	var renameTo string
	if len(args) > 0 {
		renameTo = args[0]
	} else {
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", pkgRename, eval.RenameFrom)
		var toResult interface{}
		if config.RenamePrefill {
			err := c.v.Call("input", &toResult, askMessage, eval.RenameFrom)
			if err != nil {
				return nvim.EchohlErr(c.v, pkgRename, "Keyboard interrupt")
			}
		} else {
			err := c.v.Call("input", &toResult, askMessage)
			if err != nil {
				return nvim.EchohlErr(c.v, pkgRename, "Keyboard interrupt")
			}
		}
		if toResult.(string) == "" {
			return nvim.EchohlErr(c.v, pkgRename, "Not enough arguments for rename destination name")
		}
		renameTo = fmt.Sprintf("%s", toResult)
	}

	c.v.Command(fmt.Sprintf("echo '%s: Renaming ' | echohl Identifier | echon '%s' | echohl None | echon ' to ' | echohl Identifier | echon '%s' | echohl None | echon ' ...'", pkgRename, eval.RenameFrom, renameTo))

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
			loclist, _ := quickfix.ParseError(er, eval.Cwd, &c.ctxt.Build)
			quickfix.SetLoclist(c.v, loclist)
			quickfix.OpenLoclist(c.v, w, loclist, true)
		}()

		err = errors.Wrap(err, pkgRename)
		return nvim.ErrorWrap(c.v, err)
	}

	write.Close()
	out, _ := ioutil.ReadAll(read)
	defer nvim.EchoSuccess(c.v, pkgRename, fmt.Sprintf("%s", out))

	// TODO(zchee): 'edit' command is ugly.
	// Should create tempfile and use SetBufferLines.
	return c.v.Command("silent edit")
}
