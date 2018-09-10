// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"path/filepath"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

var (
	runTerm     *nvimutil.Terminal
	runLastArgs []string
)

func (c *Command) cmdRun(args []string, file string) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Run(c.ctx, args, file)
	}()

	select {
	case <-c.ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Run", e)
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

func (c *Command) cmdRunLast(file string) {
	if len(runLastArgs) == 0 {
		err := errors.New("not found GoRun last arguments")
		nvimutil.ErrorWrap(c.Nvim, err)
		return
	}

	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Run(c.ctx, runLastArgs, file)
	}()

	select {
	case <-c.ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Run", e)
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

// Run runs the go run command for current buffer's packages.
func (c *Command) Run(ctx context.Context, args []string, file string) error {
	defer nvimutil.Profile(ctx, time.Now(), "Run")

	ctx, span := trace.StartSpan(ctx, "Run")
	defer span.End()

	cmd := []string{"go", "run", file}
	if len(args) != 0 {
		runLastArgs = args
		cmd = append(cmd, args...)
	}

	if runTerm == nil {
		runTerm = nvimutil.NewTerminal(c.Nvim, "__GO_RUN__", cmd, config.TerminalMode)
	}
	runTerm.Dir = fs.FindVCSRoot(filepath.Dir(file))

	if err := runTerm.Run(cmd); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
