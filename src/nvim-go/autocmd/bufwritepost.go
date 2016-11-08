// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"
	"nvim-go/nvimutil"
	"nvim-go/nvimutil/quickfix"

	vim "github.com/neovim/go-client/nvim"
)

type bufWritePostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (a *Autocmd) cmdBufWritePost(v *vim.Nvim, eval *bufWritePostEval) {
	go a.bufWritePost(v, eval)
}

func (a *Autocmd) bufWritePost(v *vim.Nvim, eval *bufWritePostEval) error {
	if config.FmtAutosave {
		err := <-a.bufWritePreChan
		switch e := err.(type) {
		case error:
			if e != nil {
				// normal errros
				return nvimutil.ErrorWrap(v, e)
			}
		case []*vim.QuickfixError:
			// Cleanup Errlist
			a.ctxt.Errlist = make(map[string][]*vim.QuickfixError)
			a.ctxt.Errlist["Fmt"] = e
			return quickfix.ErrorList(v, a.ctxt.Errlist, true)
		}
	}

	if config.BuildAutosave {
		err := a.c.Build(config.BuildForce, &commands.CmdBuildEval{
			Cwd: eval.Cwd,
			Dir: eval.Dir,
		})

		switch e := err.(type) {
		case error:
			// normal errros
			if e != nil {
				return nvimutil.ErrorWrap(v, e)
			}
		case []*vim.QuickfixError:
			// Cleanup Errlist
			a.ctxt.Errlist = make(map[string][]*vim.QuickfixError)
			a.ctxt.Errlist["Build"] = e
			return quickfix.ErrorList(v, a.ctxt.Errlist, true)
		}
		if len(a.ctxt.Errlist) == 0 {
			quickfix.CloseLoclist(v)
		}
	}

	if config.GoVetAutosave {
		go func() {
			// Cleanup old results
			a.ctxt.Errlist["Vet"] = nil

			errlist, err := a.c.Vet(nil, &commands.CmdVetEval{
				Cwd: eval.Cwd,
				Dir: eval.Dir,
			})
			if err != nil {
				// normal errros
				nvimutil.ErrorWrap(v, err)
				return
			}
			if errlist != nil {
				a.ctxt.Errlist["Vet"] = errlist
				if len(a.ctxt.Errlist) > 0 {
					quickfix.ErrorList(v, a.ctxt.Errlist, true)
					return
				}
			}
			if a.ctxt.Errlist["Vet"] == nil {
				quickfix.ClearErrorlist(v, true)
			}
		}()
	}

	if config.MetalinterAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.c.Metalinter(eval.Cwd)
		}()
	}

	if config.TestAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.c.Test([]string{}, eval.Dir)
		}()
	}

	a.wg.Wait()
	return nil
}
