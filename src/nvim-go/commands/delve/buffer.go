// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"fmt"

	"nvim-go/config"
	"nvim-go/nvim"

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
		d.buffer = make(map[string]*nvim.Buf)
		nnoremap := make(map[string]string)

		d.buffer[Terminal] = nvim.NewBuffer(d.v)
		d.buffer[Terminal].Create(Terminal, nvim.FiletypeDelve, fmt.Sprintf("silent belowright %d vsplit", (width*2/5)), option)
		nnoremap["i"] = fmt.Sprintf(":<C-u>call rpcrequest(%d, 'DlvStdin')<CR>", config.ChannelID)
		d.buffer[Terminal].SetLocalMapping(nvim.NoremapNormal, nnoremap)

		d.buffer[Context] = nvim.NewBuffer(d.v)
		d.buffer[Context].Create(Context, nvim.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*2/3)), option)

		d.buffer[Threads] = nvim.NewBuffer(d.v)
		d.buffer[Threads].Create(Threads, nvim.FiletypeDelve, fmt.Sprintf("silent belowright %d split", (height*1/5)), option)
		d.v.SetWindowOption(d.buffer[Threads].Window, "winfixheight", true)

		defer d.v.SetCurrentWindow(d.cw)
	}()

	var err error
	d.pcSign, err = nvim.NewSign(d.v, "delve_pc", nvim.ProgramCounterSymbol, "delvePCSign", "delvePCLine") // *nvim.Sign
	if err != nil {
		return errors.Wrap(err, "delve/createDebugBuffer")
	}

	return d.p.Wait()
}

func (d *Delve) setTerminalOption() map[nvim.NvimOption]map[string]interface{} {
	option := make(map[nvim.NvimOption]map[string]interface{})
	bufoption := make(map[string]interface{})
	bufvar := make(map[string]interface{})
	windowoption := make(map[string]interface{})

	bufoption[nvim.BufOptionBufhidden] = nvim.BufhiddenDelete
	bufoption[nvim.BufOptionBuflisted] = false
	bufoption[nvim.BufOptionBuftype] = nvim.BuftypeNofile
	bufoption[nvim.BufOptionFiletype] = nvim.FiletypeDelve
	bufoption[nvim.BufOptionModifiable] = false
	bufoption[nvim.BufOptionSwapfile] = false

	bufvar[nvim.BufVarColorcolumn] = ""

	windowoption[nvim.WinOptionList] = false
	windowoption[nvim.WinOptionNumber] = false
	windowoption[nvim.WinOptionRelativenumber] = false
	windowoption[nvim.WinOptionWinfixheight] = false

	option[nvim.BufferOption] = bufoption
	option[nvim.BufferVar] = bufvar
	option[nvim.WindowOption] = windowoption

	return option
}
