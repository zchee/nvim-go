// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"fmt"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

const (
	Terminal = "terminal"
	Context  = "context"
	Threads  = "thread"
)

func (d *delve) createDebugBuffer(v *vim.Vim) error {
	p := v.NewPipeline()

	p.CurrentBuffer(&d.cb)
	p.CurrentWindow(&d.cw)
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "delve/createDebugBuffer")
	}

	var height, width int
	p.WindowHeight(d.cw, &height)
	p.WindowWidth(d.cw, &width)
	if err := p.Wait(); err != nil {
		return errors.Annotate(err, "delve/createDebugBuffer")
	}

	bufOption := d.setNvimOption("buffer")
	bufVar := d.setNvimVar("buffer")
	winOption := d.setNvimOption("window")

	go func() {
		d.buffer = make(map[string]*buffer.Buffer)
		nnoremap := make(map[string]string)

		d.buffer[Terminal] = buffer.NewBuffer(Terminal, fmt.Sprintf("silent belowright %d vsplit", (width*2/5)), 0)
		d.buffer[Terminal].Create(v, bufOption, bufVar, winOption, nil)
		nnoremap["i"] = fmt.Sprintf(":<C-u>call rpcrequest(%d, 'DlvStdin')<CR>", config.ChannelID)
		d.buffer[Terminal].SetMapping(v, buffer.NoremapNormal, nnoremap)

		d.buffer[Context] = buffer.NewBuffer(Context, fmt.Sprintf("silent belowright %d split", (height*2/3)), 0)
		d.buffer[Context].Create(v, bufOption, bufVar, winOption, nil)

		d.buffer[Threads] = buffer.NewBuffer(Threads, fmt.Sprintf("silent belowright %d split", (height*1/5)), 0)
		d.buffer[Threads].Create(v, bufOption, bufVar, winOption, nil)

		v.SetWindowOption(d.buffer[Threads].Window, "winfixheight", true)

		defer v.SetCurrentWindow(d.cw)
	}()

	var err error
	d.pcSign, err = nvim.NewSign(v, "delve_pc", nvim.ProgramCounterSymbol, "delvePCSign", "delvePCLine") // *nvim.Sign
	if err != nil {
		return errors.Annotate(err, "delve/createDebugBuffer")
	}

	return p.Wait()
}

func (d *delve) setNvimOption(scope string) map[string]interface{} {
	options := make(map[string]interface{})

	switch scope {
	case "buffer":
		options[buffer.Bufhidden] = buffer.BufhiddenDelete
		options[buffer.Buflisted] = false
		options[buffer.Buftype] = buffer.BuftypeNofile
		options[buffer.Filetype] = buffer.FiletypeDelve
		options[buffer.OpModifiable] = false
		options[buffer.Swapfile] = false
	case "window":
		options[buffer.List] = false
		options[buffer.Number] = false
		options[buffer.Relativenumber] = false
		options[buffer.Winfixheight] = false
	}

	return options
}

func (d *delve) setNvimVar(scope string) map[string]interface{} {
	vars := make(map[string]interface{})

	switch scope {
	case "buffer":
		vars[buffer.Colorcolumn] = ""
	}

	return vars
}
