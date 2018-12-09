// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"context"
	"path/filepath"

	"github.com/neovim/go-client/nvim"

	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

type bufWritePostEval struct {
	Cwd  string `eval:"getcwd()"`
	File string `eval:"expand('%:p')"`
}

// BufWritePost run the 'autosave' commands on BufWritePost autocmd.
func (a *Autocmd) BufWritePost(pctx context.Context, eval *bufWritePostEval) error {
	ctx, span := monitoring.StartSpan(pctx, "BufWritePost")
	defer span.End()

	dir := filepath.Dir(eval.File)

	if config.FmtAutosave {
		err := <-a.bufWritePreChan
		switch e := err.(type) {
		case error:
			return nvimutil.ErrorWrap(a.Nvim, e)
		case []*nvim.QuickfixError:
			errlist := make(map[string][]*nvim.QuickfixError)
			errlist["Fmt"] = e
			return nvimutil.ErrorList(a.Nvim, errlist, true)
		}
	}

	if config.BuildAutosave {
		err := a.cmd.Build(ctx, nil, config.BuildForce, &command.CmdBuildEval{
			Cwd:  eval.Cwd,
			File: eval.File,
		})
		switch e := err.(type) {
		case error:
			return nvimutil.ErrorWrap(a.Nvim, e)
		case []*nvim.QuickfixError:
			errlist := make(map[string][]*nvim.QuickfixError)
			errlist["Build"] = e
			return nvimutil.ErrorList(a.Nvim, errlist, true)
		}
	}

	if config.GolintAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()

			a.errs.Delete("Lint")
			err := a.cmd.Lint(ctx, nil, eval.File)
			switch e := err.(type) {
			case error:
				nvimutil.ErrorWrap(a.Nvim, e)
			case []*nvim.QuickfixError:
				errlist := make(map[string][]*nvim.QuickfixError)
				errlist["Lint"] = e
				nvimutil.ErrorList(a.Nvim, errlist, true)
			}
		}()
	}

	if config.GoVetAutosave {
		a.wg.Add(1)
		a.mu.Lock()
		go func() {
			defer func() {
				a.wg.Done()
				a.mu.Unlock()
			}()

			a.errs.Delete("Vet")
			err := a.cmd.Vet(ctx, nil, &command.CmdVetEval{
				Cwd:  eval.Cwd,
				File: eval.File,
			})
			switch e := err.(type) {
			case error:
				nvimutil.ErrorWrap(a.Nvim, e)
			case []*nvim.QuickfixError:
				a.errs.Store("Vet", e)
			}
		}()
	}

	if config.MetalinterAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()

			a.errs.Delete("MetaLinter")
			err := a.cmd.Metalinter(ctx, eval.Cwd)
			switch e := err.(type) {
			case error:
				nvimutil.ErrorWrap(a.Nvim, e)
			case nil:
			}
		}()
	}

	if config.TestAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()

			a.errs.Delete("Test")
			err := a.cmd.Test(ctx, nil, dir)
			switch e := err.(type) {
			case error:
				nvimutil.ErrorWrap(a.Nvim, e)
			case nil:
			}
		}()
	}

	a.wg.Wait()
	errlist := make(map[string][]*nvim.QuickfixError)
	a.errs.Range(func(ki, vi interface{}) bool {
		k, v := ki.(string), vi.([]*nvim.QuickfixError)
		errlist[k] = append(errlist[k], v...)
		return true
	})

	if len(errlist) > 0 {
		return nvimutil.ErrorList(a.Nvim, errlist, true)
	}

	return nvimutil.ClearErrorlist(a.Nvim, true)
}
