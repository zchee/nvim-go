// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"
	"nvim-go/nvimutil"
	"path/filepath"

	"github.com/neovim/go-client/nvim"
)

type bufWritePostEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (a *Autocmd) BufWritePost(eval *bufWritePostEval) {
	go a.bufWritePost(eval)
}

func (a *Autocmd) bufWritePost(eval *bufWritePostEval) error {
	dir := filepath.Dir(eval.File)

	if config.FmtAutosave {
		err := <-a.bufWritePreChan
		switch e := err.(type) {
		case error:
			if e != nil {
				// normal errros
				return nvimutil.ErrorWrap(a.Nvim, e)
			}
		case []*nvim.QuickfixError:
			// Cleanup Errlist
			a.ctx.Errlist = make(map[string][]*nvim.QuickfixError)
			a.ctx.Errlist["Fmt"] = e
			return nvimutil.ErrorList(a.Nvim, a.ctx.Errlist, true)
		}
	}

	if config.BuildAutosave {
		err := a.cmds.Build(config.BuildForce, &commands.CmdBuildEval{
			Cwd:  eval.Cwd,
			File: eval.File,
		})

		switch e := err.(type) {
		case error:
			// normal errros
			if e != nil {
				return nvimutil.ErrorWrap(a.Nvim, e)
			}
		case []*nvim.QuickfixError:
			// Cleanup Errlist
			a.ctx.Errlist = make(map[string][]*nvim.QuickfixError)
			a.ctx.Errlist["Build"] = e
			return nvimutil.ErrorList(a.Nvim, a.ctx.Errlist, true)
		}
	}

	if config.GolintAutosave {
		a.wg.Add(1)
		a.mu.Lock()
		go func() {
			defer func() {
				a.wg.Done()
				a.mu.Unlock()
			}()

			// Cleanup old results
			a.ctx.Errlist["Lint"] = nil

			errlist, err := a.cmds.Lint(nil, eval.File)
			if err != nil {
				// normal errros
				nvimutil.ErrorWrap(a.Nvim, err)
				return
			}
			a.ctx.Errlist["Lint"] = errlist
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

			// Cleanup old results
			a.ctx.Errlist["Vet"] = nil

			errlist, err := a.cmds.Vet(nil, &commands.CmdVetEval{
				Cwd:  eval.Cwd,
				File: eval.File,
			})
			if err != nil {
				// normal errros
				nvimutil.ErrorWrap(a.Nvim, err)
				return
			}
			a.ctx.Errlist["Vet"] = errlist
		}()
	}

	if config.MetalinterAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.cmds.Metalinter(eval.Cwd)
		}()
	}

	if config.TestAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.cmds.Test([]string{}, dir)
		}()
	}

	a.wg.Wait()

	if len(a.ctx.Errlist) > 0 {
		return nvimutil.ErrorList(a.Nvim, a.ctx.Errlist, true)
	} else {
		nvimutil.ClearErrorlist(a.Nvim, true)
	}

	return nil
}
