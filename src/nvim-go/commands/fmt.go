// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"

	"nvim-go/gb"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/imports"
)

func init() {
	plugin.HandleCommand("Gofmt", &plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"}, Fmt)
	plugin.HandleAutocmd("BufWritePre", &plugin.AutocmdOptions{Pattern: "*.go"}, onBufWritePre)
}

var options = imports.Options{
	AllErrors: true,
	Comments:  true,
	TabIndent: true,
	TabWidth:  8,
}

func Fmt(v *vim.Vim, r [2]int, file string) error {
	defer gb.WithGoBuildForPath(file)()

	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}

	in, err := v.BufferLineSlice(b, 0, -1, true, true)
	if err != nil {
		return err
	}

	buf, err := imports.Process("", bytes.Join(in, []byte{'\n'}), &options)
	if err != nil {
		return nvim.LoclistErrors(v, b, err)
	} else {
		nvim.LoclistClose(v)
	}

	out := bytes.Split(bytes.TrimSuffix(buf, []byte{'\n'}), []byte{'\n'})

	return minUpdate(v, b, in, out)
}

type BufWritePre struct {
	Name string `msgpack:",array"`
	Cwd  string
}

func onBufWritePre(v *vim.Vim, eval *BufWritePre) error {
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

	p.Command("Gofmt")

	return p.Wait()
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
