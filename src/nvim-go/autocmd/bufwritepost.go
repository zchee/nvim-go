// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autocmd

import (
	"nvim-go/commands"
	"nvim-go/config"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
)

type bufWritePostEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (a *AutocmdContext) autocmdBufWritePost(v *vim.Vim, eval *bufWritePostEval) {
	switch <-a.bufWritePreChan {
	default:
		return
	case nil:
		if config.BuildAutosave {
			go func() {
				err := commands.Build(v, false, &commands.CmdBuildEval{
					Cwd: eval.Cwd,
					Dir: eval.Dir,
				})
				a.send(a.bufWritePostChan, err)
			}()
		}
		if config.MetalinterAutosave {
			go func() {
				err := commands.Metalinter(v, eval.Cwd)
				a.send(a.bufWritePostChan, err)
			}()
		}

		if config.TestAutosave {
			go func() {
				err := commands.Test(v, []string{}, eval.Dir)
				a.send(a.bufWritePostChan, err)
			}()
		}
	}

	err := <-a.bufWritePostChan
	if err != nil {
		nvim.ErrorWrap(v, err)
		return
	}
}
