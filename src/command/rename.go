// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/src/config"
	"github.com/zchee/nvim-go/src/nvimutil"
	"golang.org/x/tools/refactor/rename"
)

const pkgRename = "GoRename"

type cmdRenameEval struct {
	Cwd        string `msgpack:",array"`
	File       string
	RenameFrom string
}

func (c *Command) cmdRename(args []string, bang bool, eval *cmdRenameEval) {
	go func() {
		err := c.Rename(args, bang, eval)

		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.buildContext.Errlist["Rename"] = e
			nvimutil.ErrorList(c.Nvim, c.buildContext.Errlist, true)
		}
	}()
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func (c *Command) Rename(args []string, bang bool, eval *cmdRenameEval) interface{} {
	defer nvimutil.Profile(c.ctx, time.Now(), "GoRename")

	b := nvim.Buffer(c.buildContext.BufNr)
	w := nvim.Window(c.buildContext.WinID)

	offset, err := nvimutil.ByteOffset(c.Nvim, b, w)
	if err != nil {
		return errors.WithStack(err)
	}
	pos := fmt.Sprintf("%s:#%d", eval.File, offset)

	var renameTo string
	if len(args) > 0 {
		renameTo = args[0]
	} else {
		askMessage := fmt.Sprintf("%s: Rename '%s' to: ", pkgRename, eval.RenameFrom)
		var toResult interface{}
		if config.RenamePrefill {
			err := c.Nvim.Call("input", &toResult, askMessage, eval.RenameFrom)
			if err != nil {
				return errors.New("GoRename: Keyboard interrupt")
			}
		} else {
			err := c.Nvim.Call("input", &toResult, askMessage)
			if err != nil {
				return errors.New("GoRename: Keyboard interrupt")
			}
		}
		if toResult.(string) == "" {
			return nvimutil.Echoerr(c.Nvim, "GoRename: Not enough arguments for rename destination name")
		}
		renameTo = fmt.Sprintf("%s", toResult)
	}

	c.Nvim.Command(fmt.Sprintf("echo '%s: Renaming ' | echohl Identifier | echon '%s' | echohl None | echon ' to ' | echohl Identifier | echon '%s' | echohl None | echon ' ...'", pkgRename, eval.RenameFrom, renameTo))

	if bang {
		rename.Force = true
	}

	// TODO(zchee): More elegant way
	// save original stdout and stderr
	saveStdout, saveStderr := os.Stdout, os.Stderr
	read, write, _ := os.Pipe()
	// migrate stderr and stdout
	os.Stderr = os.Stdout
	os.Stderr = write
	defer func() {
		os.Stderr = saveStdout
		os.Stderr = saveStderr
	}()

	// TODO(zchee): reached race limit, dying when race build
	if err := rename.Main(&build.Default, pos, "", renameTo); err != nil {
		write.Close()
		renameErr, err := ioutil.ReadAll(read)
		if err != nil {
			return errors.WithStack(err)
		}

		loclist, _ := nvimutil.ParseError(renameErr, eval.Cwd, &c.buildContext.Build, nil)
		nvimutil.SetLoclist(c.Nvim, loclist)
		nvimutil.OpenLoclist(c.Nvim, w, loclist, true)

		return loclist
	}

	write.Close()
	out, _ := ioutil.ReadAll(read)
	defer nvimutil.EchoSuccess(c.Nvim, pkgRename, fmt.Sprintf("%s", out))

	// TODO(zchee): 'edit' command is ugly.
	// Should create tempfile and use SetBufferLines.
	return c.Nvim.Command("silent edit")
}
