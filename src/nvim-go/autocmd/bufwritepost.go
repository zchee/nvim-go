// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/command"
	"nvim-go/config"
	"nvim-go/nvimutil"
	"path/filepath"

	"github.com/neovim/go-client/nvim"
)

type bufWritePostEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

// BufWritePost run the 'autosave' commands on BufWritePost autocmd.
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
			a.ctx.Errlist["Fmt"] = e
			return nvimutil.ErrorList(a.Nvim, a.ctx.Errlist, true)
		}
	}

	if config.BuildAutosave {
		delete(a.ctx.Errlist, "Build")
		err := a.cmd.Build(config.BuildForce, &command.CmdBuildEval{
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
			delete(a.ctx.Errlist, "Lint")

			errlist, err := a.cmd.Lint(nil, eval.File)
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
			delete(a.ctx.Errlist, "Vet")

			err := a.cmd.Vet(nil, &command.CmdVetEval{
				Cwd:  eval.Cwd,
				File: eval.File,
			})
			switch e := err.(type) {
			case error:
				// normal errros
				if e != nil {
					nvimutil.ErrorWrap(a.Nvim, e)
				}
			case []*nvim.QuickfixError:
				// Cleanup Errlist
				a.ctx.Errlist = make(map[string][]*nvim.QuickfixError)
				a.ctx.Errlist["Vet"] = e
				nvimutil.ErrorList(a.Nvim, a.ctx.Errlist, true)
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
			a.cmd.Test([]string{}, dir)
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
