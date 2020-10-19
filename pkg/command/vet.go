// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/monitoring"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// CmdVetEval struct type for Eval of GoBuild command.
type CmdVetEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Command) cmdVet(ctx context.Context, args []string, eval *CmdVetEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Vet(ctx, args, eval)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Vet", e)
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

// Vet is a simple checker for static errors in Go source code use go tool vet command.
func (c *Command) Vet(pctx context.Context, args []string, eval *CmdVetEval) interface{} {
	ctx, span := monitoring.StartSpan(pctx, "Vet")
	defer span.End()

	vetCmd := exec.CommandContext(ctx, "go", "tool", "vet")
	vetCmd.Dir = eval.Cwd

	switch {
	case len(args) > 0:
		lastArg := args[len(args)-1]
		if !strings.HasPrefix(lastArg, "-") {
			switch path := filepath.Join(eval.Cwd, lastArg); {
			case args[0] == ".":
				vetCmd.Args = append(vetCmd.Args, ".")
			case fs.IsDir(path):
				eval.Cwd = path
				vetCmd.Args = append(vetCmd.Args, args[:len(args)-1]...)
			case fs.IsExist(path) && fs.IsGoFile(path):
				vetCmd.Args = append(vetCmd.Args, path)
			case filepath.Base(path) == "%":
				path = eval.File
				vetCmd.Args = append(vetCmd.Args, path)
			default:
				err := errors.New("Invalid directory path")
				span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
				return errors.WithStack(err)
			}
		} else {
			vetCmd.Args = append(vetCmd.Args, args...)
			vetCmd.Args = append(vetCmd.Args, ".")
		}
	case len(config.GoVetFlags) > 0:
		vetCmd.Args = append(vetCmd.Args, config.GoVetFlags...)
		vetCmd.Args = append(vetCmd.Args, ".")
	default:
		vetCmd.Args = append(vetCmd.Args, ".")
	}

	var stderr bytes.Buffer
	vetCmd.Stderr = &stderr

	vetErr := vetCmd.Run()
	if vetErr != nil {
		errlist, err := nvimutil.ParseError(ctx, stderr.Bytes(), eval.Cwd, &c.buildContext.Build, config.GoVetIgnore)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		}
		return errlist
	}

	return nil
}

// Core flags:
//
//   -asmdecl
//     	enable asmdecl analysis
//   -assign
//     	enable assign analysis
//   -atomic
//     	enable atomic analysis
//   -bools
//     	enable bools analysis
//   -buildtag
//     	enable buildtag analysis
//   -cgocall
//     	enable cgocall analysis
//   -composites
//     	enable composites analysis
//   -copylocks
//     	enable copylocks analysis
//   -errorsas
//     	enable errorsas analysis
//   -flags
//     	print analyzer flags in JSON
//   -httpresponse
//     	enable httpresponse analysis
//   -ifaceassert
//     	enable ifaceassert analysis
//   -json
//     	emit JSON output
//   -loopclosure
//     	enable loopclosure analysis
//   -lostcancel
//     	enable lostcancel analysis
//   -nilfunc
//     	enable nilfunc analysis
//   -printf
//     	enable printf analysis
//   -shift
//     	enable shift analysis
//   -stdmethods
//     	enable stdmethods analysis
//   -stringintconv
//     	enable stringintconv analysis
//   -structtag
//     	enable structtag analysis
//   -tests
//     	enable tests analysis
//   -unmarshal
//     	enable unmarshal analysis
//   -unreachable
//     	enable unreachable analysis
//   -unsafeptr
//     	enable unsafeptr analysis
//   -unusedresult
//     	enable unusedresult analysis
func (c *Command) cmdVetComplete(ctx context.Context, a *nvim.CommandCompletionArgs, dir string) ([]string, error) {
	complete, err := nvimutil.CompleteFiles(c.Nvim, a, dir)
	if err != nil {
		return nil, err
	}

	complete = append(complete, []string{
		"-asmdecl",
		"-assign",
		"-atomic",
		"-bools",
		"-buildtag",
		"-cgocall",
		"-composites",
		"-copylocks",
		"-copylocks",
		"-errorsas",
		"-flags",
		"-httpresponse",
		"-ifaceassert",
		"-loopclosure",
		"-lostcancel",
		"-nilfunc",
		"-printf",
		"-shift",
		"-stdmethods",
		"-stringintconv",
		"-structtag",
		"-tests",
		"-unmarshal",
		"-unreachable",
		"-unsafeptr",
		"-unusedresult",
	}...)
	return complete, nil
}
