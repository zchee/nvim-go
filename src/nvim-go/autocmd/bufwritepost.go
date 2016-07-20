// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/quickfix"

	"github.com/neovim-go/vim"
)

type bufWritePostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (a *Autocmd) cmdBufWritePost(v *vim.Vim, eval *bufWritePostEval) {
	go a.bufWritePost(v, eval)
}

func (a *Autocmd) bufWritePost(v *vim.Vim, eval *bufWritePostEval) error {
	if config.FmtAutosave {
		err := <-a.bufWritePreChan
		switch e := err.(type) {
		case error:
			if e != nil {
				// normal errros
				return nvim.ErrorWrap(v, e)
			}
		case []*vim.QuickfixError:
			a.ctxt.Errlist["Fmt"] = e
			// Cleanup GoBuild errors
			a.ctxt.Errlist["Build"] = nil
			return quickfix.ErrorList(v, a.ctxt.Errlist, true)
		}
	}

	if config.BuildAutosave {
		err := a.c.Build(false, &commands.CmdBuildEval{
			Cwd: eval.Cwd,
			Dir: eval.Dir,
		})

		switch e := err.(type) {
		case error:
			// normal errros
			if e != nil {
				nvim.ErrorWrap(v, e)
			}
		case []*vim.QuickfixError:
			a.ctxt.Errlist["Build"] = e
			return quickfix.ErrorList(v, a.ctxt.Errlist, true)
		}
		if len(a.ctxt.Errlist) == 0 {
			quickfix.CloseLoclist(v)
		}
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
