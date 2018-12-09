// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cweill/gotests/gotests/process"
	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

var generateFuncRe = regexp.MustCompile(`(?m)^func\s(?:\(\w\s[[:graph:]]+\)\s)?([\w]+)\(`)

func (c *Command) cmdGenerateTest(ctx context.Context, args []string, ranges [2]int, bang bool, dir string) {
	errch := make(chan error, 1)
	go func() {
		errch <- c.GenerateTest(ctx, args, ranges, bang, dir)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case nil:
			// nothing to do
		}
	}
}

// GenerateTest generates the test files based by current buffer or args files
// functions.
func (c *Command) GenerateTest(ctx context.Context, args []string, ranges [2]int, bang bool, dir string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	defer nvimutil.Profile(ctx, time.Now(), "GenerateTest")
	span := trace.FromContext(ctx)
	span.SetName("GenerateTest")
	defer span.End()

	b := nvim.Buffer(c.buildContext.BufNr)
	if len(args) == 0 {
		f, err := c.Nvim.BufferName(b)
		if err != nil {
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}
		args = []string{f}
	}

	opt := &process.Options{
		WriteOutput:   true,
		PrintInputs:   true,
		AllFuncs:      config.GenerateTestAllFuncs,
		ExclFuncs:     config.GenerateTestExclFuncs,
		ExportedFuncs: config.GenerateTestExportedFuncs,
		Subtests:      config.GenerateTestSubTest,
	}

	// Check users used range. range return variable: (1,$)
	// If not used visual range, always ranges[0] is 0.
	if ranges[0] != 1 {
		// Re-check range[1] is not buffer line count
		lines, err := c.Nvim.BufferLineCount(b)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}

		if ranges[1] != lines {
			start, end := ranges[0], ranges[1]
			// Get the buffer 2D slice
			// Neovim range value is based 1
			blines, err := c.Nvim.BufferLines(b, start-1, end, true)
			if err != nil {
				span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
				return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
			}
			// Convert to 1D byte slice
			buf := nvimutil.ToByteSlice(blines)

			matches := generateFuncRe.FindAllSubmatch(buf, -1)
			var onlyFuncs []string
			for _, fnName := range matches {
				onlyFuncs = append(onlyFuncs, string(fnName[1]))
			}
			opt.AllFuncs = false
			opt.ExportedFuncs = false
			// Set onlyFuncs option
			// like "-only=^(fooFunc|barFunc)$"
			opt.OnlyFuncs = fmt.Sprintf("^(%s)$", strings.Join(onlyFuncs, "|"))
		}
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	process.Run(w, args, opt)

	w.Close()
	os.Stdout = oldStdout

	if !bang {
		var genFuncs string
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			genFuncs += scan.Text() + "\n"
		}

		// TODO(zchee): More beautiful code
		suffix := "_test.go "
		var ftests, ftestsRel string
		for _, f := range args {
			fnAbs := strings.Split(f, filepath.Ext(f))
			ftests += fnAbs[0] + suffix

			_, fnRel := filepath.Split(fnAbs[0])
			ftestsRel += fnRel + suffix
		}

		ask := fmt.Sprintf("%s\nGoGenerateTest: Generated %s\nGoGenerateTest: Open it? (y, n): ", genFuncs, ftestsRel)
		var answer interface{}
		if err := c.Nvim.Call("input", &answer, ask); err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}
		// TODO(zchee): Support open the ftests[0] file only.
		// If passes multiple files for 'edit' commands, occur 'E172: Only one file name allowed' errror.
		if answer.(string) == "y" {
			return c.Nvim.Command(fmt.Sprintf("edit %s", ftests))
		}
	}

	return nil
}
