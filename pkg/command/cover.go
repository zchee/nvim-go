// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/internal/cover"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// cmdCoverEval struct type for Eval of GoBuild command.
type cmdCoverEval struct {
	Cwd  string `msgpack:",array"`
	File string `msgpack:",array"`
}

func (c *Command) cmdCover(eval *cmdCoverEval) {
	go func() {
		err := c.cover(eval)

		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Cover", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		}
	}()
}

// cover run the go tool cover command and highlight current buffer based cover
// profile result.
func (c *Command) cover(eval *cmdCoverEval) interface{} {
	defer nvimutil.Profile(c.ctx, time.Now(), "GoCover")

	coverFile, err := ioutil.TempFile(os.TempDir(), "nvim-go-cover")
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.Remove(coverFile.Name())

	cmd := exec.Command("go", strings.Fields(fmt.Sprintf("test -cover -covermode=%s -coverprofile=%s .", config.CoverMode, coverFile.Name()))...)
	if len(config.CoverFlags) > 0 {
		cmd.Args = append(cmd.Args, config.CoverFlags...)
	}
	cmd.Dir = filepath.Dir(eval.File)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if coverErr := cmd.Run(); coverErr != nil && coverErr.(*exec.ExitError) != nil {
		errlist, err := nvimutil.ParseError(stdout.Bytes(), filepath.Dir(eval.File), &c.buildContext.Build, nil)
		if err != nil {
			return errors.WithStack(err)
		}
		return errlist
	}
	delete(c.buildContext.Errlist, "Cover")

	profile, err := cover.ParseProfiles(coverFile.Name())
	if err != nil {
		return errors.WithStack(err)
	}

	b, err := c.Nvim.CurrentBuffer()
	if err != nil {
		return errors.WithStack(err)
	}
	buf, err := c.Nvim.BufferLines(b, 0, -1, true)
	if err != nil {
		return errors.WithStack(err)
	}

	highlighted := make(map[int]bool)
	var res int // for ignore the msgpack decode errror. not used
	batch := c.Nvim.NewBatch()
	for _, prof := range profile {
		if filepath.Base(prof.FileName) == filepath.Base(eval.File) {

			if config.DebugEnable {
				log.Printf("prof.Blocks:\n%+v\n", spew.Sdump(prof.Blocks))
				log.Printf("prof.Boundaries():\n%+v\n", spew.Sdump(prof.Boundaries(nvimutil.ToByteSlice(buf))))
			}
			for _, block := range prof.Blocks {
				for line := block.StartLine - 1; line <= block.EndLine-1; line++ { // nvim_buf_add_highlight line started by 0
					// not highlighting the last RBRACE of the function
					if line == block.EndLine-1 && block.EndCol == 2 {
						break
					}

					var hl string
					switch {
					case block.Count == 0:
						hl = "GoCoverMiss"
					case block.Count-block.NumStmt == 0:
						hl = "GoCoverPartial"
					default:
						hl = "GoCoverHit"
					}
					if !highlighted[line] {
						batch.AddBufferHighlight(b, 0, hl, line, 0, -1, &res)
						highlighted[line] = true
					}
				}
			}
		}
	}

	return errors.WithStack(batch.Execute())
}
