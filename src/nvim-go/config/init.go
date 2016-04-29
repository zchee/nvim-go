package config

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

// Config struct of config variable for nvim-go commands.
type Config struct {
	AstView    AstViewVars
	Build      BuildVars
	Fmt        FmtVars
	Guru       GuruVars
	Iferr      IferrVars
	Metalinter MetalinterVars
	Terminal   TerminalVars
	Debug      DebugVars
}

// AstVars GoAstView command config variable.
type AstViewVars struct {
	FoldIcon string `eval:"g:go#ast#foldicon"`
}

// BuildVars GoBuild command config variable.
type BuildVars struct {
	Autosave int64 `eval:"g:go#build#autosave"`
}

// FmtVars GoFmt command config variable.
type FmtVars struct {
	Async int64 `eval:"g:go#fmt#async"`
}

// GuruVars GoGuru command config variable.
type GuruVars struct {
	Reflection int64 `eval:"g:go#guru#reflection"`
	KeepCursor int64 `eval:"g:go#guru#keep_cursor"`
	JumpFirst  int64 `eval:"g:go#guru#jump_first"`
}

// IferrVars GoIferr command config variable.
type IferrVars struct {
	IferrAutosave int64 `eval:"g:go#iferr#autosave"`
}

// MetalinterVars GoMetaLinter command config variable.
type MetalinterVars struct {
	Autosave      int64    `eval:"g:go#lint#metalinter#autosave"`
	AutosaveTools []string `eval:"g:go#lint#metalinter#autosave#tools"`
	Tools         []string `eval:"g:go#lint#metalinter#tools"`
	Deadline      string   `eval:"g:go#lint#metalinter#deadline"`
}

// TerminalVars configure of open the terminal window
type TerminalVars struct {
	Mode         string `eval:"g:go#terminal#mode"`
	Position     string `eval:"g:go#terminal#position"`
	Height       int64  `eval:"g:go#terminal#height"`
	Width        int64  `eval:"g:go#terminal#width"`
	StartInsetrt int64  `eval:"g:go#terminal#start_insert"`
}

// DebugVars debug of nvim-go config variable.
type DebugVars struct {
	Pprof int64 `eval:"g:go#debug#pprof"`
}

func init() {
	plugin.HandleAutocmd("VimEnter",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "*"}, Getconfig)
}

var (
	// AstFoldIcon define default astview tree fold icon.
	AstFoldIcon string
	// BuildAutosave call the GoBuild command automatically at during the BufWritePost.
	BuildAutosave bool
	// FmtAsync asynchronous call the GoFmt command at during the BufWritePre.
	FmtAsync bool
	// GuruReflection use the type reflection on GoGuru commmands.
	GuruReflection bool
	// GuruKeepCursor keep the cursor focus to source buffer instead of quickfix or locationlist.
	GuruKeepCursor bool
	// GuruJumpFirst jump the first error position on GoGuru commands.
	GuruJumpFirst bool
	// IferrAutosave call the GoIferr command automatically at during the BufWritePre.
	IferrAutosave bool
	// MetalinterAutosave call the GoMetaLinter command automatically at during the BufWritePre.
	MetalinterAutosave bool
	// MetalinterAutosaveTools lint tool list for MetalinterAutosave.
	MetalinterAutosaveTools []string
	// MetalinterTools lint tool list for GoMetaLinter command.
	MetalinterTools []string
	// MetalinterDeadline deadline of GoMetaLinter command timeout.
	MetalinterDeadline string
	// TerminalMode open the terminal window mode.
	TerminalMode string
	// TerminalPosition open the terminal window position.
	TerminalPosition string
	// TerminalHeight open the terminal window height.
	TerminalHeight int64
	// TerminalWidth open the terminal window width.
	TerminalWidth int64
	// TerminalStartInsert workaround if users set "autocmd BufEnter term://* startinsert"
	TerminalStartInsert bool
	// DebugPprof Enable net/http/pprof debugging.
	DebugPprof bool
)

// Getconfig define the user config variables to Go global varialble.
func Getconfig(v *vim.Vim, cfg *Config) {
	// AstView
	AstFoldIcon = cfg.AstView.FoldIcon

	// Build
	BuildAutosave = itob(cfg.Build.Autosave)

	// Fmt
	FmtAsync = itob(cfg.Fmt.Async)

	// Guru
	GuruReflection = itob(cfg.Guru.Reflection)
	GuruKeepCursor = itob(cfg.Guru.KeepCursor)
	GuruJumpFirst = itob(cfg.Guru.JumpFirst)

	// Iferr
	IferrAutosave = itob(cfg.Iferr.IferrAutosave)

	// Metalinter
	MetalinterAutosave = itob(cfg.Metalinter.Autosave)
	MetalinterAutosaveTools = cfg.Metalinter.AutosaveTools
	MetalinterTools = cfg.Metalinter.Tools
	MetalinterDeadline = cfg.Metalinter.Deadline

	// Terminal
	TerminalMode = cfg.Terminal.Mode
	TerminalPosition = cfg.Terminal.Position
	TerminalHeight = cfg.Terminal.Height
	TerminalWidth = cfg.Terminal.Width
	TerminalStartInsert = itob(cfg.Terminal.StartInsetrt)

	// Debug
	DebugPprof = itob(cfg.Debug.Pprof)
}

func itob(i int64) bool {
	if i == int64(0) {
		return false
	}
	return true
}
