// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"go/scanner"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
	"golang.org/x/tools/imports"
)

const pkgFmt = "GoFmt"

var importsOptions = imports.Options{
	AllErrors: true,
	Comments:  true,
	TabIndent: true,
	TabWidth:  8,
}

func init() {
	plugin.HandleCommand("Gofmt", &plugin.CommandOptions{Eval: "expand('%:p:h')"}, Fmt)
}

// Fmt format to the current buffer source uses gofmt behavior.
func Fmt(v *vim.Vim, dir string) error {
	defer profile.Start(time.Now(), pkgFmt)
	ctxt := new(context.Context)
	defer ctxt.Build.SetContext(dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	bufName, err := v.BufferName(b)
	if err != nil {
		return err
	}

	in, err := v.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}

	buf, err := imports.Process("", nvim.ToByteSlice(in), &importsOptions)
	if err != nil {
		var loclist []*quickfix.ErrorlistData

		if e, ok := err.(scanner.Error); ok {
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: bufName,
				LNum:     e.Pos.Line,
				Col:      e.Pos.Column,
				Text:     e.Msg,
			})
		} else if el, ok := err.(scanner.ErrorList); ok {
			for _, e := range el {
				loclist = append(loclist, &quickfix.ErrorlistData{
					FileName: bufName,
					LNum:     e.Pos.Line,
					Col:      e.Pos.Column,
					Text:     e.Msg,
				})
			}
		}

		if err := quickfix.SetLoclist(v, loclist); err != nil {
			return nvim.Echomsg(v, "Gofmt:", err)
		}

		quickfix.OpenLoclist(v, w, loclist, true)
		return errors.Annotate(err, pkgFmt)
	}

	out := nvim.ToBufferLines(bytes.TrimSuffix(buf, []byte{'\n'}))

	minUpdate(v, b, in, out)

	// TODO(zchee): When executed Fmt(itself) function at autocmd BufWritePre, vim "write"
	// command will starting before the finish of the Fmt function because that function called
	// asynchronously uses goroutine.
	// However, "noautocmd" raises duplicate the filesystem events.
	// In the case of macOS fsevents:
	//  (FSE_STAT_CHANGED -> FSE_CHOWN -> FSE_CONTENT_MODIFIED) x2.
	// It will affect the watchdog system such as inotify-tools, fswatch or fsevents-tools.
	// We need to consider the Alternative of BufWriteCmd or other an effective way.
	return v.Command("noautocmd write")
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
