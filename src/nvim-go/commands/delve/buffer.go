// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"fmt"
	"time"

	"nvim-go/config"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

func (d *delve) createDebugBuffer(v *vim.Vim, p *vim.Pipeline) error {
	defer profile.Start(time.Now(), "delve/createDebugBuffer")

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

	d.buffers = make([]*buffer.Buffer, 4, 5)
	for i, n := range []string{"terminal", "threads", "stacktrace", "locals"} {
		d.buffers[i] = buffer.NewBuffer(n)
	}
	d.buffers[0].Size = (width * 2 / 5)
	d.buffers[1].Size = (height * 2 / 3)
	d.buffers[2].Size = (width * 1 / 5)
	d.buffers[3].Size = (height * 1 / 2)
	// d.buffers[4].Size = (height * 2 / 5)
	d.buffers[0].Mode = fmt.Sprintf("silent belowright %d vsplit", d.buffers[0].Size)
	d.buffers[1].Mode = fmt.Sprintf("silent belowright %d split", d.buffers[1].Size)
	d.buffers[2].Mode = fmt.Sprintf("silent belowright %d vsplit", d.buffers[2].Size)
	d.buffers[3].Mode = fmt.Sprintf("silent belowright %d split", d.buffers[3].Size)
	// d.buffers[4].Mode = fmt.Sprintf("silent belowright %d split", d.buffers[4].Size)

	for _, buf := range d.buffers {
		if err := buf.Create(v, bufOption, bufVar, winOption, nil); err != nil {
			return errors.Annotate(err, "delve/createDebugBuffer")
		}
	}

	nnoremap := make(map[string]string)
	nnoremap["i"] = fmt.Sprintf(":<C-u>call rpcrequest(%d, 'DlvStdin')<CR>", config.ChannelID)
	d.buffers[0].SetMapping(v, buffer.NoremapNormal, nnoremap)

	var err error
	d.pcSign, err = nvim.NewSign(v, "delve_pc", nvim.ProgramCounterSymbol, "delvePCSign", "delvePCLine") // *nvim.Sign
	if err != nil {
		return errors.Annotate(err, "delve/createDebugBuffer")
	}

	p.SetCurrentWindow(d.cw)
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
		options[buffer.Modifiable] = false
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
	options := make(map[string]interface{})

	switch scope {
	case "buffer":
		options[buffer.Colorcolumn] = ""
	}

	return options
}
