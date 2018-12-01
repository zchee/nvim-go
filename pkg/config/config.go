// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"github.com/neovim/go-client/nvim"
)

// Config represents a config variable for nvim-go.
// Each type must be exported for plugin.HandleAutocmd Eval option.
// Also it does not support embeded type.
type Config struct {
	Global *Global

	Build    *build
	Cover    *cover
	Fmt      *fmt
	Generate *generate
	Guru     *guru
	Iferr    *iferr
	Lint     *lint
	Rename   *rename
	Terminal *terminal
	Test     *test

	Debug *debug
}

// Global represents a global config variable.
type Global struct {
	ChannelID     int
	ServerName    string `eval:"v:servername"`
	ErrorListType string `eval:"get(g:, 'go#global#errorlisttype', 'locationlist')"`
}

// build GoBuild command config variable.
type build struct {
	Appengine int64    `eval:"get(g:, 'go#build#appengine', 0)"`
	Autosave  int64    `eval:"get(g:, 'go#build#autosave', 0)"`
	Force     int64    `eval:"get(g:, 'go#build#force', 0)"`
	Flags     []string `eval:"get(g:, 'go#build#flags', [])"`
	IsNotGb   int64    `eval:"get(g:, 'go#build#is_not_gb', 0)"`
}

type cover struct {
	Flags []string `eval:"get(g:, 'go#cover#flags', [])"`
	Mode  string   `eval:"get(g:, 'go#cover#mode', 'atomic')"`
}

// fmt represents a GoFmt command config variable.
type fmt struct {
	Autosave       int64    `eval:"get(g:, 'go#fmt#autosave', 0)"`
	Mode           string   `eval:"get(g:, 'go#fmt#mode', 'goimports')"`
	GoImportsLocal []string `eval:"get(g:, 'go#fmt#goimports_local', [])"`
}

// generate represents a GoGenerate command config variables.
type generate struct {
	TestAllFuncs      int64  `eval:"get(g:, 'go#generate#test#allfuncs', 1)"`
	TestExclFuncs     string `eval:"get(g:, 'go#generate#test#exclude', '')"`
	TestExportedFuncs int64  `eval:"get(g:, 'go#generate#test#exportedfuncs', 0)"`
	TestSubTest       int64  `eval:"get(g:, 'go#generate#test#subtest', 1)"`
}

// guru represents a GoGuru command config variable.
type guru struct {
	Reflection int64            `eval:"get(g:, 'go#guru#reflection', 0)"`
	KeepCursor map[string]int64 `eval:"get(g:, 'go#guru#keep_cursor', {'callees':0,'callers':0,'callstack':0,'definition':0,'describe':0,'freevars':0,'implements':0,'peers':0,'pointsto':0,'referrers':0,'whicherrs':0})"`
	JumpFirst  int64            `eval:"get(g:, 'go#guru#jump_first', 0)"`
}

// iferr represents a GoIferr command config variable.
type iferr struct {
	Autosave int64 `eval:"get(g:, 'go#iferr#autosave', 0)"`
}

// lint represents a code lint commands config variable.
type lint struct {
	GolintAutosave          int64    `eval:"get(g:, 'go#lint#golint#autosave', 0)"`
	GolintIgnore            []string `eval:"get(g:, 'go#lint#golint#ignore', [])"`
	GolintMinConfidence     float64  `eval:"get(g:, 'go#lint#golint#min_confidence', 0.8)"`
	GolintMode              string   `eval:"get(g:, 'go#lint#golint#mode', 'current')"`
	GoVetAutosave           int64    `eval:"get(g:, 'go#lint#govet#autosave', 0)"`
	GoVetFlags              []string `eval:"get(g:, 'go#lint#govet#flags', [])"`
	GoVetIgnore             []string `eval:"get(g:, 'go#lint#govet#ignore', [])"`
	MetalinterAutosave      int64    `eval:"get(g:, 'go#lint#metalinter#autosave', 0)"`
	MetalinterAutosaveTools []string `eval:"get(g:, 'go#lint#metalinter#autosave#tools', ['vet', 'golint'])"`
	MetalinterTools         []string `eval:"get(g:, 'go#lint#metalinter#tools', ['vet', 'golint'])"`
	MetalinterDeadline      string   `eval:"get(g:, 'go#lint#metalinter#deadline', '5s')"`
	MetalinterSkipDir       []string `eval:"get(g:, 'go#lint#metalinter#skip_dir', [])"`
}

// rename represents a GoRename command config variable.
type rename struct {
	Prefill int64 `eval:"get(g:, 'go#rename#prefill', 0)"`
}

// terminal represents a configure of Neovim terminal buffer.
type terminal struct {
	Mode       string `eval:"get(g:, 'go#terminal#mode', 'vsplit')"`
	Position   string `eval:"get(g:, 'go#terminal#position', 'belowright')"`
	Height     int64  `eval:"get(g:, 'go#terminal#height', 0)"`
	Width      int64  `eval:"get(g:, 'go#terminal#width', 0)"`
	StopInsert int64  `eval:"get(g:, 'go#terminal#stop_insert', 1)"`
}

// Test represents a GoTest command config variables.
type test struct {
	AllPackage int64    `eval:"get(g:, 'go#test#all_package', 0)"`
	Autosave   int64    `eval:"get(g:, 'go#test#autosave', 0)"`
	Flags      []string `eval:"get(g:, 'go#test#flags', [])"`
}

// Debug represents a debug of nvim-go config variable.
type debug struct {
	Enable int64 `eval:"get(g:, 'go#debug', 0)"`
	Pprof  int64 `eval:"get(g:, 'go#debug#pprof', 0)"`
}

var (
	// ChannelID remote plugins channel id.
	ChannelID int
	// ServerName Neovim socket listen location.
	ServerName string
	// ErrorListType type of error list window.
	ErrorListType string

	// BuildAppengine enable appengine bulid.
	BuildAppengine bool
	// BuildAutosave call the GoBuild command automatically at during the BufWritePost.
	BuildAutosave bool
	// BuildForce builds the binary instead of fake(use ioutil.TempFiile) build.
	BuildForce bool
	// BuildFlags flag of compile tools build command.
	BuildFlags []string

	// BuildIsNotGb workaround for not ues gb compiler.
	BuildIsNotGb bool

	// CoverFlags flags for cover command.
	CoverFlags []string
	// CoverMode mode of cover command.
	CoverMode string

	// FmtAutosave call the GoFmt command automatically at during the BufWritePre.
	FmtAutosave bool
	// FmtMode formatting mode of Fmt command.
	FmtMode string
	// FmtGoImportsLocal list packages of goimports -local flag.
	FmtGoImportsLocal []string

	// GenerateTestAllFuncs accept all functions to the GenerateTest.
	GenerateTestAllFuncs bool
	// GenerateTestExclFuncs exclude function of GenerateTest.
	GenerateTestExclFuncs string
	// GenerateTestExportedFuncs accept exported functions to the GenerateTest.
	GenerateTestExportedFuncs bool
	// GenerateTestSubTest whether the use Go subtest idiom or not.
	GenerateTestSubTest bool

	// GuruReflection use the type reflection on GoGuru commmands.
	GuruReflection bool
	// GuruKeepCursor keep the cursor focus to source buffer instead of quickfix or locationlist.
	GuruKeepCursor map[string]int64
	// GuruJumpFirst jump the first error position on GoGuru commands.
	GuruJumpFirst bool

	// IferrAutosave call the GoIferr command automatically at during the BufWritePre.
	IferrAutosave bool

	// GolintAutosave call the GoLint command automatically at during the BufWritePost.
	GolintAutosave bool
	// GolintIgnore ignore file for lint command.
	GolintIgnore []string
	// GolintMinConfidence minimum confidence of a problem to print it
	GolintMinConfidence float64
	// GolintMode mode of golint. available value are "root", "current" and "recursive".
	GolintMode string
	// GoVetAutosave call the GoVet command automatically at during the BufWritePost.
	GoVetAutosave bool
	// GoVetFlags default flags for GoVet commands
	GoVetFlags []string
	// GoVetIgnore ignore directories for go vet command.
	GoVetIgnore []string
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
	// TerminalStopInsert workaround if users set "autocmd BufEnter term://* startinsert".
	TerminalStopInsert bool

	// TestAutosave call the GoBuild command automatically at during the BufWritePost.
	TestAutosave bool
	// TestAll enable all package test on GoTest. similar "go test ./...", but ignored vendor and testdata.
	TestAll bool
	// TestFlags test command default flags.
	TestFlags []string

	// DebugEnable Enable debugging.
	DebugEnable bool
	// DebugPprof Enable net/http/pprof debugging.
	DebugPprof bool
)

// Get gets the user config variables and convert to global varialble.
func Get(v *nvim.Nvim, cfg *Config) {
	// Client
	ChannelID = cfg.Global.ChannelID
	ServerName = cfg.Global.ServerName
	ErrorListType = cfg.Global.ErrorListType

	// Build
	BuildAppengine = itob(cfg.Build.Appengine)
	BuildAutosave = itob(cfg.Build.Autosave)
	BuildForce = itob(cfg.Build.Force)
	BuildFlags = cfg.Build.Flags
	BuildIsNotGb = itob(cfg.Build.IsNotGb)

	// Cover
	CoverFlags = cfg.Cover.Flags
	CoverMode = cfg.Cover.Mode

	// Fmt
	FmtAutosave = itob(cfg.Fmt.Autosave)
	FmtMode = cfg.Fmt.Mode
	FmtGoImportsLocal = cfg.Fmt.GoImportsLocal

	// Generate
	GenerateTestAllFuncs = itob(cfg.Generate.TestAllFuncs)
	GenerateTestExclFuncs = cfg.Generate.TestExclFuncs
	GenerateTestExportedFuncs = itob(cfg.Generate.TestExportedFuncs)
	GenerateTestSubTest = itob(cfg.Generate.TestSubTest)

	// Guru
	GuruReflection = itob(cfg.Guru.Reflection)
	GuruKeepCursor = cfg.Guru.KeepCursor
	GuruJumpFirst = itob(cfg.Guru.JumpFirst)

	// Iferr
	IferrAutosave = itob(cfg.Iferr.Autosave)

	// Lint
	GolintAutosave = itob(cfg.Lint.GolintAutosave)
	GolintIgnore = cfg.Lint.GolintIgnore
	GolintMinConfidence = cfg.Lint.GolintMinConfidence
	GolintMode = cfg.Lint.GolintMode
	GoVetAutosave = itob(cfg.Lint.GoVetAutosave)
	GoVetFlags = cfg.Lint.GoVetFlags
	GoVetIgnore = cfg.Lint.GoVetIgnore
	MetalinterAutosave = itob(cfg.Lint.MetalinterAutosave)
	MetalinterAutosaveTools = cfg.Lint.MetalinterAutosaveTools
	MetalinterTools = cfg.Lint.MetalinterTools
	MetalinterDeadline = cfg.Lint.MetalinterDeadline
	MetalinterSkipDir = cfg.Lint.MetalinterSkipDir

	// Rename
	RenamePrefill = itob(cfg.Rename.Prefill)

	// Terminal
	TerminalMode = cfg.Terminal.Mode
	TerminalPosition = cfg.Terminal.Position
	TerminalHeight = cfg.Terminal.Height
	TerminalWidth = cfg.Terminal.Width
	TerminalStopInsert = itob(cfg.Terminal.StopInsert)

	// Test
	TestAutosave = itob(cfg.Test.Autosave)
	TestAll = itob(cfg.Test.AllPackage)
	TestFlags = cfg.Test.Flags

	// Debug
	DebugEnable = itob(cfg.Debug.Enable)
	DebugPprof = itob(cfg.Debug.Pprof)
}

func itob(i int64) bool { return i != int64(0) }
