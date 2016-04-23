package config

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

// Config struct of config variable for nvim-go commands.
type Config struct {
	Build      BuildVars
	Fmt        FmtVars
	Guru       GuruVars
	Iferr      IferrVars
	Metalinter MetalinterVars
	Debug      DebugVars
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

// DebugVars debug of nvim-go config variable.
type DebugVars struct {
	Pprof int64 `eval:"g:go#debug#pprof"`
}

func init() {
	plugin.HandleAutocmd("VimEnter",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "*"}, Getconfig)
}

var (
	// BuildAutosave call the GoBuild command automatically at during the BufWritePost.
	BuildAutosave int64
	// FmtAsync asynchronous call the GoFmt command at during the BufWritePre.
	FmtAsync int64
	// GuruReflection use the type reflection on GoGuru commmands.
	GuruReflection int64
	// GuruKeepCursor keep the cursor focus to source buffer instead of quickfix or locationlist.
	GuruKeepCursor int64
	// GuruJumpFirst jump the first error position on GoGuru commands.
	GuruJumpFirst int64
	// IferrAutosave call the GoIferr command automatically at during the BufWritePre.
	IferrAutosave int64
	// MetalinterAutosave call the GoMetaLinter command automatically at during the BufWritePre.
	MetalinterAutosave int64
	// MetalinterAutosaveTools lint tool list for MetalinterAutosave.
	MetalinterAutosaveTools []string
	// MetalinterTools lint tool list for GoMetaLinter command.
	MetalinterTools []string
	// MetalinterDeadline deadline of GoMetaLinter command timeout.
	MetalinterDeadline string
	// DebugPprof Enable net/http/pprof debugging.
	DebugPprof int64
)

// Getconfig define the user config variables to Go global varialble.
func Getconfig(v *vim.Vim, cfg *Config) {
	// Build
	BuildAutosave = cfg.Build.Autosave

	// Fmt
	FmtAsync = cfg.Fmt.Async

	// Guru
	GuruReflection = cfg.Guru.Reflection
	GuruKeepCursor = cfg.Guru.KeepCursor
	GuruJumpFirst = cfg.Guru.JumpFirst

	// Iferr
	IferrAutosave = cfg.Iferr.IferrAutosave

	// Metalinter
	MetalinterAutosave = cfg.Metalinter.Autosave
	MetalinterAutosaveTools = cfg.Metalinter.AutosaveTools
	MetalinterTools = cfg.Metalinter.Tools
	MetalinterDeadline = cfg.Metalinter.Deadline

	// Debug
	DebugPprof = cfg.Debug.Pprof
}
