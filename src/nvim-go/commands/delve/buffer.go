// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"fmt"

	"nvim-go/config"
	"nvim-go/nvimutil"

	"github.com/pkg/errors"
)

const (
	// Terminal define terminal buffer name.
	Terminal nvimutil.BufferName = "terminal"
	// Context define context buffer name.
	Context nvimutil.BufferName = "context"
	// Threads define threads buffer name.
	Threads nvimutil.BufferName = "thread"
)

func (d *Delve) createDebugBuffer() error {
	d.Batch.CurrentBuffer(&d.cb)
	d.Batch.CurrentWindow(&d.cw)
	err := d.Batch.Execute()
	if err != nil {
		return errors.WithStack(err)
	}

	var height, width int
	d.Batch.WindowHeight(d.cw, &height)
	d.Batch.WindowWidth(d.cw, &width)
	err = d.Batch.Execute()
	if err != nil {
		return errors.WithStack(err)
	}

	go func() {
		defer d.Nvim.SetCurrentWindow(d.cw)

		option := d.setTerminalOption()
		d.buffer = make(map[nvimutil.BufferName]*nvimutil.Buffer)
		nnoremap := make(map[string]string)

		d.buffer[Terminal] = nvimutil.NewBuffer(d.Nvim)
		d.buffer[Terminal].Create(string(Terminal), nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d vsplit", (width*2/5)), option)
		nnoremap["i"] = fmt.Sprintf(":<C-u>call rpcrequest(%d, 'DlvStdin')<CR>", config.ChannelID)
		d.buffer[Terminal].SetLocalMapping(nvimutil.NoremapNormal, nnoremap)

		d.buffer[Context] = nvimutil.NewBuffer(d.Nvim)
		d.buffer[Context].Create(string(Context), nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*2/3)), option)

		d.buffer[Threads] = nvimutil.NewBuffer(d.Nvim)
		d.buffer[Threads].Create(string(Threads), nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*1/5)), option)
		d.Nvim.SetWindowOption(d.buffer[Threads].Window, "winfixheight", true)

	}()

	d.pcSign, err = nvimutil.NewSign(d.Nvim, "delve_pc", nvimutil.ProgramCounterSymbol, "delvePCSign", "delvePCLine") // *nvim.Sign
	if err != nil {
		return errors.WithStack(err)
	}

	return d.Batch.Execute()
}

func (d *Delve) setTerminalOption() map[nvimutil.NvimOption]map[string]interface{} {
	option := make(map[nvimutil.NvimOption]map[string]interface{})
	bufoption := make(map[string]interface{})
	bufvar := make(map[string]interface{})
	windowoption := make(map[string]interface{})

	bufoption[nvimutil.BufOptionBufhidden] = nvimutil.BufhiddenDelete
	bufoption[nvimutil.BufOptionBuflisted] = false
	bufoption[nvimutil.BufOptionBuftype] = nvimutil.BuftypeNofile
	bufoption[nvimutil.BufOptionFiletype] = nvimutil.FiletypeDelve
	bufoption[nvimutil.BufOptionModifiable] = false
	bufoption[nvimutil.BufOptionSwapfile] = false

	bufvar[nvimutil.BufVarColorcolumn] = ""

	windowoption[nvimutil.WinOptionList] = false
	windowoption[nvimutil.WinOptionNumber] = false
	windowoption[nvimutil.WinOptionRelativenumber] = false
	windowoption[nvimutil.WinOptionWinfixheight] = false

	option[nvimutil.BufferOption] = bufoption
	option[nvimutil.BufferVar] = bufvar
	option[nvimutil.WindowOption] = windowoption

	return option
}
