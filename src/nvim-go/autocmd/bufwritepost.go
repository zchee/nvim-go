// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
)

type bufWritePostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (a *Autocmd) bufWritePost(v *vim.Vim, eval *bufWritePostEval) {
	select {
	case err := <-a.bufWritePreChan:
		if err != nil {
			return
		}
	}

	if config.BuildAutosave {
		err := commands.Build(v, false, &commands.CmdBuildEval{
			Cwd: eval.Cwd,
			Dir: eval.Dir,
		})
		if err != nil {
			return
		}
	}

	if config.MetalinterAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			commands.Metalinter(v, eval.Cwd)
		}()
	}

	if config.TestAutosave {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			commands.Test(v, []string{}, eval.Dir)
		}()
	}

	a.wg.Wait()
}
