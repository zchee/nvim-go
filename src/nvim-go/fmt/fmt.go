package fmt

import (
	"bytes"
	"nvim-go/util"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/imports"
)

func init() {
	plugin.HandleCommand("Gofmt", &plugin.CommandOptions{Range: "%", Eval: "expand('%:p')"}, fmt)
}

var options = imports.Options{
	AllErrors: true,
	Comments:  true,
	TabIndent: true,
	TabWidth:  8,
}

func fmt(v *vim.Vim, r [2]int, file string) error {
	defer util.WithGoBuildForPath(file)()

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
		return util.ReportErrors(v, b, err)
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
