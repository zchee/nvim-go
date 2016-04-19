package vars

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

type Vars struct {
	Build      BuildVars
	Fmt        FmtVars
	Guru       GuruVars
	Iferr      IferrVars
	Metalinter MetalinterVars
	Debug      DebugVars
}

type BuildVars struct {
	Autosave int64 `eval:"g:go#build#autosave"`
}

type FmtVars struct {
	Async int64 `eval:"g:go#fmt#async"`
}

type GuruVars struct {
	Reflection  int64 `eval:"g:go#guru#reflection"`
	KeepCursor  int64 `eval:"g:go#guru#keep_cursor"`
	JumpToError int64 `eval:"g:go#guru#jump_to_error"`
}

type IferrVars struct {
	IferrAutosave int64 `eval:"g:go#iferr#autosave"`
}

type MetalinterVars struct {
	Autosave      int64    `eval:"g:go#lint#metalinter#autosave"`
	AutosaveTools []string `eval:"g:go#lint#metalinter#autosave#tools"`
	Tools         []string `eval:"g:go#lint#metalinter#tools"`
	Deadline      string   `eval:"g:go#lint#metalinter#deadline"`
}

type DebugVars struct {
	Pprof int64 `eval:"g:go#debug#pprof"`
}

func init() {
	plugin.HandleAutocmd("VimEnter",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "*"}, GetVars)
}

var (
	BuildAutosave           int64
	FmtAsync                int64
	GuruReflection          int64
	GuruKeepCursor          int64
	GuruJumpToError         int64
	IferrAutosave           int64
	MetalinterAutosave      int64
	MetalinterAutosaveTools []string
	MetalinterTools         []string
	MetalinterDeadline      string
	DebugPprof              int64
)

func GetVars(v *vim.Vim, vars *Vars) {
	// Build
	BuildAutosave = vars.Build.Autosave

	// Fmt
	FmtAsync = vars.Fmt.Async

	// Guru
	GuruReflection = vars.Guru.Reflection
	GuruKeepCursor = vars.Guru.KeepCursor
	GuruJumpToError = vars.Guru.JumpToError

	// Iferr
	IferrAutosave = vars.Iferr.IferrAutosave

	// Metalinter
	MetalinterAutosave = vars.Metalinter.Autosave
	MetalinterAutosaveTools = vars.Metalinter.AutosaveTools
	MetalinterTools = vars.Metalinter.Tools
	MetalinterDeadline = vars.Metalinter.Deadline

	// Debug
	DebugPprof = vars.Debug.Pprof
}
