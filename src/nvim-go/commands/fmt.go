// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"go/scanner"
	"time"

	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/juju/errors"
	"github.com/neovim-go/vim"
	"golang.org/x/tools/imports"
)

const pkgFmt = "GoFmt"

var importsOptions = imports.Options{
	AllErrors: true,
	Comments:  true,
	TabIndent: true,
	TabWidth:  8,
}

func (c *Commands) cmdFmt(dir string) {
	go c.Fmt(dir)
}

// Fmt format to the current buffer source uses gofmt behavior.
func (c *Commands) Fmt(dir string) error {
	defer profile.Start(time.Now(), pkgFmt)
	defer c.ctxt.Build.SetContext(dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	if c.p == nil {
		c.p = c.v.NewPipeline()
	}
	c.p.CurrentBuffer(&b)
	c.p.CurrentWindow(&w)
	if err := c.p.Wait(); err != nil {
		return err
	}

	bufName, err := c.v.BufferName(b)
	if err != nil {
		return err
	}

	in, err := c.v.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}

	buf, err := imports.Process("", nvim.ToByteSlice(in), &importsOptions)
	if err != nil {
		var errlist []*vim.QuickfixError

		if e, ok := err.(scanner.Error); ok {
			errlist = append(errlist, &vim.QuickfixError{
				FileName: bufName,
				LNum:     e.Pos.Line,
				Col:      e.Pos.Column,
				Text:     e.Msg,
			})
		} else if el, ok := err.(scanner.ErrorList); ok {
			for _, e := range el {
				errlist = append(errlist, &vim.QuickfixError{
					FileName: bufName,
					LNum:     e.Pos.Line,
					Col:      e.Pos.Column,
					Text:     e.Msg,
				})
			}
		}
		c.errlist["Fmt"] = errlist

		quickfix.ErrorList(c.v, w, quickfix.LocationList, c.errlist, true)
		return errors.Annotate(err, pkgFmt)
	}

	delete(c.errlist, "Fmt")
	defer quickfix.ErrorList(c.v, w, quickfix.LocationList, c.errlist, true)

	out := nvim.ToBufferLines(bytes.TrimSuffix(buf, []byte{'\n'}))

	minUpdate(c.v, b, in, out)

	// TODO(zchee): When executed Fmt(itself) function at autocmd BufWritePre, vim "write"
	// command will starting before the finish of the Fmt function because that function called
	// asynchronously uses goroutine.
	// However, "noautocmd" raises duplicate the filesystem events.
	// In the case of macOS fsevents:
	//  (FSE_STAT_CHANGED -> FSE_CHOWN -> FSE_CONTENT_MODIFIED) x2.
	// It will affect the watchdog system such as inotify-tools, fswatch or fsevents-tools.
	// We need to consider the Alternative of BufWriteCmd or other an effective way.
	return c.v.Command("noautocmd write")
}

func minUpdate(v *vim.Vim, b vim.Buffer, in [][]byte, out [][]byte) error {
	// Find matching head lines.
	n := len(out)
	if len(in) < len(out) {
		n = len(in)
	}
	head := 0
	for ; head < n; head++ {
		if !bytes.Equal(in[head], out[head]) {
			break
		}
	}

	// Nothing to do?
	if head == len(in) && head == len(out) {
		return nil
	}

	// Find matching tail lines.
	n -= head
	tail := 0
	for ; tail < n; tail++ {
		if !bytes.Equal(in[len(in)-tail-1], out[len(out)-tail-1]) {
			break
		}
	}

	// Update the buffer.
	start := head
	end := len(in) - tail
	repl := out[head : len(out)-tail]

	return v.SetBufferLines(b, start, end, true, repl)
}
