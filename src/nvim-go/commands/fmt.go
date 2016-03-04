// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"go/scanner"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/imports"

	"nvim-go/gb"
	"nvim-go/nvim"
)

var (
	fmtAsync  = "go#fmt#async"
	vFmtAsync interface{}
)

var options = imports.Options{
	AllErrors: true,
	Comments:  true,
	TabIndent: true,
	TabWidth:  8,
}

func init() {
	plugin.HandleCommand("Gofmt", &plugin.CommandOptions{Range: "%", Eval: "expand('%:p:h')"}, Fmt)
	plugin.HandleAutocmd("BufWritePre", &plugin.AutocmdOptions{Pattern: "*.go", Eval: "expand('%:p:h')"}, onBufWritePre)
}

func onBufWritePre(v *vim.Vim, dir string) error {
	v.Var(fmtAsync, &vFmtAsync)
	if vFmtAsync.(int64) == int64(1) {
		go Fmt(v, [2]int{0, 0}, dir)
		return nil
	} else {
		return Fmt(v, [2]int{0, 0}, dir)
	}
}

func Fmt(v *vim.Vim, r [2]int, dir string) error {
	defer gb.WithGoBuildForPath(dir)()

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

	in, err := v.BufferLineSlice(b, 0, -1, true, true)
	if err != nil {
		return err
	}

	buf, err := imports.Process("", bytes.Join(in, []byte{'\n'}), &options)
	if err != nil {
		var loclist []*nvim.ErrorlistData

		if e, ok := err.(scanner.Error); ok {
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: bufName,
				LNum:     e.Pos.Line,
				Col:      e.Pos.Column,
				Text:     e.Msg,
			})
		} else if el, ok := err.(scanner.ErrorList); ok {
			for _, e := range el {
				loclist = append(loclist, &nvim.ErrorlistData{
					FileName: bufName,
					LNum:     e.Pos.Line,
					Col:      e.Pos.Column,
					Text:     e.Msg,
				})
			}
		}
		if err := nvim.SetLoclist(p, loclist); err != nil {
			return nvim.Echomsg(v, "Gofmt: %v", err)
		}
		return nvim.OpeLoclist(p, w, false)
	} else {
		nvim.CloseLoclist(v)
	}

	out := bytes.Split(bytes.TrimSuffix(buf, []byte{'\n'}), []byte{'\n'})

	return minUpdate(v, b, in, out)
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
	includeStart := true
	start := head
	end := len(in) - tail
	repl := out[head : len(out)-tail]

	if start == len(in) {
		start = -1
		includeStart = false
	}

	return v.SetBufferLineSlice(b, start, end, includeStart, false, repl)
}
