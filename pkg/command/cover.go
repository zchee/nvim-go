// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"golang.org/x/tools/cover"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// cmdCoverEval struct type for Eval of GoBuild command.
type cmdCoverEval struct {
	Cwd  string `msgpack:",array"`
	File string `msgpack:",array"`
}

func (c *Command) cmdCover(ctx context.Context, eval *cmdCoverEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.cover(ctx, eval)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errch:
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
		case nil:
			// nothing to do
		}
	}
}

// cover run the go tool cover command and highlight current buffer based cover
// profile result.
func (c *Command) cover(pctx context.Context, eval *cmdCoverEval) interface{} {
	ctx, span := monitoring.StartSpan(pctx, "Cover")
	defer span.End()

	coverFile, err := ioutil.TempFile(os.TempDir(), "nvim-go-cover")
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	defer os.Remove(coverFile.Name())

	cmd := exec.CommandContext(ctx, "go", strings.Fields(fmt.Sprintf("test -cover -covermode=atomic -coverpkg=./... -coverprofile=%s .", coverFile.Name()))...)
	if len(config.CoverFlags) > 0 {
		cmd.Args = append(cmd.Args, config.CoverFlags...)
	}
	cmd.Dir = filepath.Dir(eval.File)
	logger.FromContext(ctx).Debug("cover", zap.Any("cmd", cmd))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if coverErr := cmd.Run(); coverErr != nil && coverErr.(*exec.ExitError) != nil {
		errlist, err := nvimutil.ParseError(ctx, stdout.Bytes(), filepath.Dir(eval.File), &c.buildContext.Build, nil)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		}
		return errlist
	}
	delete(c.buildContext.Errlist, "Cover")

	profile, err := cover.ParseProfiles(coverFile.Name())
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}

	b, err := c.Nvim.CurrentBuffer()
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}

	highlighted := make(map[int]bool)
	var res int // for ignore the msgpack decode errror. not used
	batch := c.Nvim.NewBatch()
	for _, prof := range profile {
		if filepath.Base(prof.FileName) != filepath.Base(eval.File) {
			continue
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

	if err := batch.Execute(); err != nil {
		batchErr, ok := err.(*nvim.BatchError)
		if ok {
			err = batchErr.Err
		}
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}

	return nil
}
