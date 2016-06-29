// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import "github.com/garyburd/neovim-go/vim"

// Config represents a config variable for nvim-go.
// Each type must be exported for plugin.HandleAutocmd Eval option.
// Also it does not support embeded type.
type Config struct {
	Client     Client
	Analyze    analyze
	Build      build
	Fmt        fmt
	Generate   generate
	Guru       guru
	Iferr      iferr
	Metalinter metalinter
	Rename     rename
	Terminal   terminal
	Test       test
	Debug      debug
}

// Client represents a Neovim remote client information.
type Client struct {
	ChannelID  int
	ServerName string `eval:"v:servername"`
}

// analyze represents a GoAnalyze command config variable.
type analyze struct {
	FoldIcon string `eval:"g:go#analyze#foldicon"`
}

// build GoBuild command config variable.
type build struct {
	Autosave int64 `eval:"g:go#build#autosave"`
	Force    int64 `eval:"g:go#build#force"`
}

// fmt represents a GoFmt command config variable.
type fmt struct {
	Async int64 `eval:"g:go#fmt#async"`
}

// generate represents a GoGenerate command config variables.
type generate struct {
	ExclFuncs string `eval:"g:go#generate#exclude"`
}

// guru represents a GoGuru command config variable.
type guru struct {
	Reflection int64            `eval:"g:go#guru#reflection"`
	KeepCursor map[string]int64 `eval:"g:go#guru#keep_cursor"`
	JumpFirst  int64            `eval:"g:go#guru#jump_first"`
}

// iferr represents a GoIferr command config variable.
type iferr struct {
	Autosave int64 `eval:"g:go#iferr#autosave"`
}

// metalinter represents a GoMetaLinter command config variable.
type metalinter struct {
	Autosave      int64    `eval:"g:go#lint#metalinter#autosave"`
	AutosaveTools []string `eval:"g:go#lint#metalinter#autosave#tools"`
	Tools         []string `eval:"g:go#lint#metalinter#tools"`
	Deadline      string   `eval:"g:go#lint#metalinter#deadline"`
	SkipDir       []string `eval:"g:go#lint#metalinter#skip_dir"`
}

// rename represents a GoRename command config variable.
type rename struct {
	Prefill int64 `eval:"g:go#rename#prefill"`
}

// terminal represents a configure of Neovim terminal buffer.
type terminal struct {
	Mode         string `eval:"g:go#terminal#mode"`
	Position     string `eval:"g:go#terminal#position"`
	Height       int64  `eval:"g:go#terminal#height"`
	Width        int64  `eval:"g:go#terminal#width"`
	StartInsetrt int64  `eval:"g:go#terminal#start_insert"`
}

// Test represents a GoTest command config variables.
type test struct {
	Autosave int64    `eval:"g:go#test#autosave"`
	Args     []string `eval:"g:go#test#args"`
}

// Debug represents a debug of nvim-go config variable.
type debug struct {
	Pprof int64 `eval:"g:go#debug#pprof"`
}

var (
	// ClientChannelID remote plugins channel id.
	ClientChannelID int
	// ClientServerName Neovim socket listen location.
	ClientServerName string

	// AnalyzeFoldIcon define default astview tree fold icon.
	AnalyzeFoldIcon string

	// BuildAutosave call the GoBuild command automatically at during the BufWritePost.
	BuildAutosave bool
	// BuildForce builds the binary instead of fake(use ioutil.TempFiile) build.
	BuildForce bool

	// FmtAsync asynchronous call the GoFmt command at during the BufWritePre.
	FmtAsync bool

	// GenerateExclFuncs exclude function of generate test.
	GenerateExclFuncs string

	// GuruReflection use the type reflection on GoGuru commmands.
	GuruReflection bool
	// GuruKeepCursor keep the cursor focus to source buffer instead of quickfix or locationlist.
	GuruKeepCursor map[string]int64
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
	// MetalinterSkipDir skips of lint of the directory.
	MetalinterSkipDir []string

	// RenamePrefill Enable naming prefill.
	RenamePrefill bool

	// TerminalMode open the terminal window mode.
	TerminalMode string
	// TerminalPosition open the terminal window position.
	TerminalPosition string
	// TerminalHeight open the terminal window height.
	TerminalHeight int64
	// TerminalWidth open the terminal window width.
	TerminalWidth int64
	// TerminalStartInsert workaround if users set "autocmd BufEnter term://* startinsert".
	TerminalStartInsert bool

	// TestAutosave call the GoBuild command automatically at during the BufWritePost.
	TestAutosave bool
	// TestArgs test command default args.
	TestArgs []string

	// DebugPprof Enable net/http/pprof debugging.
	DebugPprof bool
)

// GetConfig gets the user config variables and convert to global varialble.
func GetConfig(v *vim.Vim, cfg *Config) {
	// Client
	ClientChannelID = cfg.Client.ChannelID
	ClientServerName = cfg.Client.ServerName

	// Analyze
	AnalyzeFoldIcon = cfg.Analyze.FoldIcon

	// Build
	BuildAutosave = itob(cfg.Build.Autosave)
	BuildForce = itob(cfg.Build.Force)

	// Fmt
	FmtAsync = itob(cfg.Fmt.Async)

	// Generate
	GenerateExclFuncs = cfg.Generate.ExclFuncs

	// Guru
	GuruReflection = itob(cfg.Guru.Reflection)
	GuruKeepCursor = cfg.Guru.KeepCursor
	GuruJumpFirst = itob(cfg.Guru.JumpFirst)

	// Iferr
	IferrAutosave = itob(cfg.Iferr.Autosave)

	// Metalinter
	MetalinterAutosave = itob(cfg.Metalinter.Autosave)
	MetalinterAutosaveTools = cfg.Metalinter.AutosaveTools
	MetalinterTools = cfg.Metalinter.Tools
	MetalinterDeadline = cfg.Metalinter.Deadline
	MetalinterSkipDir = cfg.Metalinter.SkipDir

	// Rename
	RenamePrefill = itob(cfg.Rename.Prefill)

	// Terminal
	TerminalMode = cfg.Terminal.Mode
	TerminalPosition = cfg.Terminal.Position
	TerminalHeight = cfg.Terminal.Height
	TerminalWidth = cfg.Terminal.Width
	TerminalStartInsert = itob(cfg.Terminal.StartInsetrt)

	// Test
	TestAutosave = itob(cfg.Test.Autosave)
	TestArgs = cfg.Test.Args

	// Debug
	DebugPprof = itob(cfg.Debug.Pprof)
}

func itob(i int64) bool { return i != int64(0) }
