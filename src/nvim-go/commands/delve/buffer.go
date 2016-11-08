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
	Terminal = "terminal"
	// Context define context buffer name.
	Context = "context"
	// Threads define threads buffer name.
	Threads = "thread"
)

func (d *Delve) createDebugBuffer() error {
	d.p.CurrentBuffer(&d.cb)
	d.p.CurrentWindow(&d.cw)
	if err := d.p.Wait(); err != nil {
		return errors.Wrap(err, "delve/createDebugBuffer")
	}

	var height, width int
	d.p.WindowHeight(d.cw, &height)
	d.p.WindowWidth(d.cw, &width)
	if err := d.p.Wait(); err != nil {
		return errors.Wrap(err, "delve/createDebugBuffer")
	}

	go func() {
		option := d.setTerminalOption()
		d.buffer = make(map[string]*nvimutil.Buf)
		nnoremap := make(map[string]string)

		d.buffer[Terminal] = nvimutil.NewBuffer(d.v)
		d.buffer[Terminal].Create(Terminal, nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d vsplit", (width*2/5)), option)
		nnoremap["i"] = fmt.Sprintf(":<C-u>call rpcrequest(%d, 'DlvStdin')<CR>", config.ChannelID)
		d.buffer[Terminal].SetLocalMapping(nvimutil.NoremapNormal, nnoremap)

		d.buffer[Context] = nvimutil.NewBuffer(d.v)
		d.buffer[Context].Create(Context, nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*2/3)), option)

		d.buffer[Threads] = nvimutil.NewBuffer(d.v)
		d.buffer[Threads].Create(Threads, nvimutil.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*1/5)), option)
		d.v.SetWindowOption(d.buffer[Threads].Window, "winfixheight", true)

		defer d.v.SetCurrentWindow(d.cw)
	}()

	var err error
	d.pcSign, err = nvimutil.NewSign(d.v, "delve_pc", nvimutil.ProgramCounterSymbol, "delvePCSign", "delvePCLine") // *nvim.Sign
	if err != nil {
		return errors.Wrap(err, "delve/createDebugBuffer")
	}

	return d.p.Wait()
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
