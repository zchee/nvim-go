// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

// NvimOption represents a Neovim buffer, window and tabpage options.
type NvimOption int

const (
	// BufferOption buffer option type.
	BufferOption NvimOption = iota
	// BufferVar buffer var type.
	BufferVar
	// WindowOption window option type.
	WindowOption
	// WindowVar window var type.
	WindowVar
	// TabpageVar tabpage var type.
	TabpageVar
)

const (
	// BufOptionBufhidden represents a bufhidden.
	BufOptionBufhidden = "bufhidden" // string
	// BufOptionBuflisted represents a buflisted.
	BufOptionBuflisted = "buflisted" // bool
	// BufOptionBuftype represents a buftype.
	BufOptionBuftype = "buftype" // string
	// BufOptionFiletype represents a filetype.
	BufOptionFiletype = "filetype" // string
	// BufOptionModifiable represents a modifiable.
	BufOptionModifiable = "modifiable" // bool
	// BufOptionModified represents a modified.
	BufOptionModified = "modified" // bool
	// BufOptionSwapfile represents a swapfile.
	BufOptionSwapfile = "swapfile" // bool

	// BufVarColorcolumn represents a colorcolumn.
	BufVarColorcolumn = "colorcolumn" // string

	// WinOptionList represents a list.
	WinOptionList = "list" // bool
	// WinOptionNumber represents a number.
	WinOptionNumber = "number" // bool
	// WinOptionRelativenumber represents a relativenumbers.
	WinOptionRelativenumber = "relativenumber" // bool
	// WinOptionWinfixheight represents a winfixheight.
	WinOptionWinfixheight = "winfixheight" // bool
)

const (
	// BufhiddenDelete delete the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bdelete.
	BufhiddenDelete = "delete"
	// BufhiddenHide hide the buffer (don't unload it), also when 'hidden' is not set.
	BufhiddenHide = "hide"
	// BufhiddenUnload unload the buffer, also when 'hidden' is set or using :hide.
	BufhiddenUnload = "unload"
	// BufhiddenWipe wipe out the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bwipeout.
	BufhiddenWipe = "wipe"
	// BuftypeAcwrite buffer which will always be written with BufWriteCmd autocommands.
	BuftypeAcwrite = "acwrite"
	// BuftypeHelp help buffer (you are not supposed to set this manually).
	BuftypeHelp = "help"
	// BuftypeNofile buffer which is not related to a file and will not be written.
	BuftypeNofile = "nofile"
	// BuftypeNowrite buffer which will not be written.
	BuftypeNowrite = "nowrite"
	// BuftypeQuickfix quickfix buffer, contains list of errors :cwindow or list of locations :lwindow.
	BuftypeQuickfix = "quickfix"
	// BuftypeTerminal terminal buffer, this is set automatically when a terminal is created. See nvim-terminal-emulator for more information.
	BuftypeTerminal = "terminal"
)

const (
	// FiletypeAsm represents a asm filetype.
	FiletypeAsm = "asm"
	// FiletypeC represents a c filetype.
	FiletypeC = "c"
	// FiletypeCpp represents a cpp filetype.
	FiletypeCpp = "cpp"
	// FiletypeDelve represents a delve filetype.
	FiletypeDelve = "delve"
	// FiletypeGas represents a gas filetype.
	FiletypeGas = "gas"
	// FiletypeGo represents a go filetype.
	FiletypeGo = "go"
	// FiletypeSh represents a sh filetype.
	FiletypeSh = "sh"
	// FiletypeTerminal represents a terminal filetype.
	FiletypeTerminal = "terminal"
	// FiletypeGoTerminal represents a go-terminal filetype.
	FiletypeGoTerminal = "goterminal"
)
