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
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// cmdCoverEval struct type for Eval of GoBuild command.
type cmdCoverEval struct {
	Cwd  string `msgpack:",array"`
	File string `msgpack:",array"`
}

func (c *Command) cmdCover(ctx context.Context, args []string, eval *cmdCoverEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.cover(ctx, args, eval)
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
func (c *Command) cover(pctx context.Context, args []string, eval *cmdCoverEval) interface{} {
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
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args...)
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

	profiles, err := cover.ParseProfiles(coverFile.Name())
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}

	buffer, err := c.Nvim.CurrentBuffer()
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	nsID, err := c.Nvim.CreateNamespace("nvim-go")
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	c.namespaceID = nsID

	batch := c.Nvim.NewBatch()
	var res int
	highlighted := make(map[int]bool)

	for _, profile := range profiles {
		if filepath.Base(profile.FileName) != filepath.Base(eval.File) {
			continue
		}

		for _, block := range profile.Blocks {
			for line := block.StartLine - 1; line <= block.EndLine-1; line++ { // nvim_buf_add_highlight line is started by 0
				if line == block.EndLine-1 && block.EndCol == 2 {
					break // not highlighting the last RBRACE of the function
				}

				if highlighted[line] {
					continue
				}

				var hl string
				switch {
				case block.Count == 0:
					hl = "GoCoverMiss"
				case block.Count-block.NumStmt == 0: // TODO(zchee): handle GoCoverPartial
					hl = "GoCoverPartial"
				default:
					hl = "GoCoverHit"
				}

				batch.AddBufferHighlight(buffer, nsID, hl, line, 0, -1, &res)
				highlighted[line] = true
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

func (c *Command) cmdClearCover(ctx context.Context) (err error) {
	if c.namespaceID == 0 {
		return
	}

	buffer, err := c.Nvim.CurrentBuffer()
	if err != nil {
		return err
	}

	err = c.Nvim.ClearBufferHighlight(buffer, c.namespaceID, 0, -1)

	return
}
