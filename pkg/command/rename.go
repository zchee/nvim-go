// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/tools/refactor/rename"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

const pkgRename = "GoRename"

type cmdRenameEval struct {
	Cwd        string `msgpack:",array"`
	File       string
	RenameFrom string
}

func (c *Command) cmdRename(ctx context.Context, args []string, bang bool, eval *cmdRenameEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Rename(ctx, args, bang, eval)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Rename", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		case nil:
			// nothing to do
		}
	}
}

// Rename rename the current cursor word use golang.org/x/tools/refactor/rename.
func (c *Command) Rename(pctx context.Context, args []string, bang bool, eval *cmdRenameEval) interface{} {
	ctx, span := monitoring.StartSpan(pctx, "Rename")
	defer span.End()

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
	rd, wd, err := os.Pipe()
	if err != nil {
		return errors.WithStack(err)
	}

	// migrate and stderr and stdout
	os.Stderr = os.Stdout
	os.Stderr = wd
	defer func() {
		os.Stderr = saveStdout
		os.Stderr = saveStderr
	}()

	bctxt := &build.Default
	logger.FromContext(ctx).Debug("Rename", zap.String("bctxt.GOROOT", bctxt.GOROOT), zap.String("bctxt.GOPATH", bctxt.GOPATH))

	// TODO(zchee): reached race limit, dying when race build
	if err = rename.Main(bctxt, pos, "", renameTo); err != nil {
		wd.Close()
		renameErr, err := ioutil.ReadAll(rd)
		if err != nil {
			return errors.WithStack(err)
		}

		loclist, _ := nvimutil.ParseError(ctx, renameErr, eval.Cwd, &c.buildContext.Build, nil)
		nvimutil.SetLoclist(c.Nvim, loclist)
		nvimutil.OpenLoclist(c.Nvim, w, loclist, true)

		return loclist
	}

	wd.Close()
	out, err := ioutil.ReadAll(rd)
	if err != nil {
		return errors.WithStack(err)
	}
	defer nvimutil.EchoSuccess(c.Nvim, pkgRename, fmt.Sprintf("%s", out))

	// TODO(zchee): 'edit' command is ugly.
	// Should create tempfile and use SetBufferLines.
	return c.Nvim.Command("silent edit")
}
