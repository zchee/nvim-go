package nvim

const (
	// Buffer options
	BufOptionBufhidden  = "bufhidden"  // string
	BufOptionBuflisted  = "buflisted"  // bool
	BufOptionBuftype    = "buftype"    // string
	BufOptionFiletype   = "filetype"   // string
	BufOptionModifiable = "modifiable" // bool
	BufOptionModified   = "modified"   // bool
	BufOptionSwapfile   = "swapfile"   // bool

	// Buffer var
	BufVarColorcolumn = "colorcolumn" // string

	// Window options
	WinOptionList           = "list"           // bool
	WinOptionNumber         = "number"         // bool
	WinOptionRelativenumber = "relativenumber" // bool
	WinOptionWinfixheight   = "winfixheight"   // bool
)

const (
	BufhiddenDelete  = "delete"   // delete the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bdelete.
	BufhiddenHide    = "hide"     // hide the buffer (don't unload it), also when 'hidden' is not set.
	BufhiddenUnload  = "unload"   // unload the buffer, also when 'hidden' is set or using :hide.
	BufhiddenWipe    = "wipe"     // wipe out the buffer from the buffer list, also when 'hidden' is set or using :hide, like using :bwipeout.
	BuftypeAcwrite   = "acwrite"  // buffer which will always be written with BufWriteCmd autocommands.
	BuftypeHelp      = "help"     // help buffer (you are not supposed to set this manually)
	BuftypeNofile    = "nofile"   // buffer which is not related to a file and will not be written.
	BuftypeNowrite   = "nowrite"  // buffer which will not be written.
	BuftypeQuickfix  = "quickfix" // quickfix buffer, contains list of errors :cwindow or list of locations :lwindow
	BuftypeTerminal  = "terminal" // terminal buffer, this is set automatically when a terminal is created. See nvim-terminal-emulator for more information.
	FiletypeAsm      = "asm"
	FiletypeC        = "c"
	FiletypeCpp      = "cpp"
	FiletypeDelve    = "delve"
	FiletypeGas      = "gas"
	FiletypeGo       = "go"
	FiletypeTerminal = "terminal"
	FiletypeAnalyze  = "goanalyze"
)
