// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"path/filepath"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/zchee/nvim-go/pkg/command"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

type bufWritePostEval struct {
	Cwd  string `eval:"getcwd()"`
	File string `eval:"expand('%:p')"`
}

func (a *Autocmd) bufWritePost(eval *bufWritePostEval) {
	go a.BufWritePost(eval)
}

// BufWritePost run the 'autosave' commands on BufWritePost autocmd.
func (a *Autocmd) BufWritePost(eval *bufWritePostEval) error {
	defer nvimutil.Profile(a.ctx, time.Now(), "BufWritePost")

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
		err := a.cmd.Build(nil, config.BuildForce, &command.CmdBuildEval{
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
			errlist, err := a.cmd.Lint(nil, eval.File)
			if err != nil {
				nvimutil.ErrorWrap(a.Nvim, err)
				return
			}
			a.errs.Store("Lint", errlist)
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
			err := a.cmd.Vet(nil, &command.CmdVetEval{
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
			a.cmd.Metalinter(eval.Cwd)
		}()
	}

	if config.TestAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.cmd.Test(nil, dir)
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
