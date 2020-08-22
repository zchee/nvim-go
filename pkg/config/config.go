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
	Appengine bool     `eval:"get(g:, 'go#build#appengine', v:false)"`
	Autosave  bool     `eval:"get(g:, 'go#build#autosave', v:false)"`
	Force     bool     `eval:"get(g:, 'go#build#force', v:false)"`
	Flags     []string `eval:"get(g:, 'go#build#flags', [])"`
	IsNotGb   bool     `eval:"get(g:, 'go#build#is_not_gb', v:false)"`
}

type cover struct {
	Flags []string `eval:"get(g:, 'go#cover#flags', [])"`
	Mode  string   `eval:"get(g:, 'go#cover#mode', 'atomic')"`
}

// fmt represents a GoFmt command config variable.
type fmt struct {
	Autosave       bool     `eval:"get(g:, 'go#fmt#autosave', v:false)"`
	Mode           string   `eval:"get(g:, 'go#fmt#mode', 'goimports')"`
	GoImportsLocal []string `eval:"get(g:, 'go#fmt#goimports_local', [])"`
}

// generate represents a GoGenerate command config variables.
type generate struct {
	TestAllFuncs       bool   `eval:"get(g:, 'go#generate#test#allfuncs', v:true)"`
	TestExclFuncs      string `eval:"get(g:, 'go#generate#test#exclude', '')"`
	TestExportedFuncs  bool   `eval:"get(g:, 'go#generate#test#exportedfuncs', v:false)"`
	TestSubTest        bool   `eval:"get(g:, 'go#generate#test#subtest', v:true)"`
	TestParallel       bool   `eval:"get(g:, 'go#generate#test#parallel', v:true)"`
	TestTemplateDir    string `eval:"get(g:, 'go#generate#test#template_dir', '')"`
	TemplateParamsPath string `eval:"get(g:, 'go#generate#test#template_params_path', '')"`
}

// guru represents a GoGuru command config variable.
type guru struct {
	Reflection bool            `eval:"get(g:, 'go#guru#reflection', v:false)"`
	KeepCursor map[string]bool `eval:"get(g:, 'go#guru#keep_cursor', {'callees':v:false,'callers':v:false,'callstack':v:false,'definition':v:false,'describe':v:false,'freevars':v:false,'implements':v:false,'peers':v:false,'pointsto':v:false,'referrers':v:false,'whicherrs':v:false})"`
	JumpFirst  bool            `eval:"get(g:, 'go#guru#jump_first', v:false)"`
}

// iferr represents a GoIferr command config variable.
type iferr struct {
	Autosave bool `eval:"get(g:, 'go#iferr#autosave', v:false)"`
}

// lint represents a code lint commands config variable.
type lint struct {
	GolintAutosave          bool     `eval:"get(g:, 'go#lint#golint#autosave', v:false)"`
	GolintIgnore            []string `eval:"get(g:, 'go#lint#golint#ignore', [])"`
	GolintMinConfidence     float64  `eval:"get(g:, 'go#lint#golint#min_confidence', 0.8)"`
	GolintMode              string   `eval:"get(g:, 'go#lint#golint#mode', 'current')"`
	GoVetAutosave           bool     `eval:"get(g:, 'go#lint#govet#autosave', v:false)"`
	GoVetFlags              []string `eval:"get(g:, 'go#lint#govet#flags', [])"`
	GoVetIgnore             []string `eval:"get(g:, 'go#lint#govet#ignore', [])"`
	MetalinterAutosave      bool     `eval:"get(g:, 'go#lint#metalinter#autosave', v:false)"`
	MetalinterAutosaveTools []string `eval:"get(g:, 'go#lint#metalinter#autosave#tools', ['vet', 'golint'])"`
	MetalinterTools         []string `eval:"get(g:, 'go#lint#metalinter#tools', ['vet', 'golint'])"`
	MetalinterDeadline      string   `eval:"get(g:, 'go#lint#metalinter#deadline', '5s')"`
	MetalinterSkipDir       []string `eval:"get(g:, 'go#lint#metalinter#skip_dir', [])"`
}

// rename represents a GoRename command config variable.
type rename struct {
	Prefill bool `eval:"get(g:, 'go#rename#prefill', v:false)"`
}

// terminal represents a configure of Neovim terminal buffer.
type terminal struct {
	Mode       string `eval:"get(g:, 'go#terminal#mode', 'vsplit')"`
	Position   string `eval:"get(g:, 'go#terminal#position', 'belowright')"`
	Height     int64  `eval:"get(g:, 'go#terminal#height', 0)"`
	Width      int64  `eval:"get(g:, 'go#terminal#width', 0)"`
	StopInsert bool   `eval:"get(g:, 'go#terminal#stop_insert', v:true)"`
}

// Test represents a GoTest command config variables.
type test struct {
	AllPackage bool     `eval:"get(g:, 'go#test#all_package', v:false)"`
	Autosave   bool     `eval:"get(g:, 'go#test#autosave', v:false)"`
	Flags      []string `eval:"get(g:, 'go#test#flags', [])"`
}

// Debug represents a debug of nvim-go config variable.
type debug struct {
	Enable bool `eval:"get(g:, 'go#debug', v:false)"`
	Pprof  bool `eval:"get(g:, 'go#debug#pprof', v:false)"`
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
	// GenerateTestParallel print tests that runs the subtests in parallel.
	GenerateTestParallel bool
	// GenerateTestTemplateDir path to custom template set.
	GenerateTestTemplateDir string
	// GenerateTestTemplateParamsPath path to custom paramters json file(s).
	GenerateTestTemplateParamsPath string

	// GuruReflection use the type reflection on GoGuru commmands.
	GuruReflection bool
	// GuruKeepCursor keep the cursor focus to source buffer instead of quickfix or locationlist.
	GuruKeepCursor map[string]bool
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
	BuildAppengine = cfg.Build.Appengine
	BuildAutosave = cfg.Build.Autosave
	BuildForce = cfg.Build.Force
	BuildFlags = cfg.Build.Flags
	BuildIsNotGb = cfg.Build.IsNotGb

	// Cover
	CoverFlags = cfg.Cover.Flags
	CoverMode = cfg.Cover.Mode

	// Fmt
	FmtAutosave = cfg.Fmt.Autosave
	FmtMode = cfg.Fmt.Mode
	FmtGoImportsLocal = cfg.Fmt.GoImportsLocal

	// Generate
	GenerateTestAllFuncs = cfg.Generate.TestAllFuncs
	GenerateTestExclFuncs = cfg.Generate.TestExclFuncs
	GenerateTestExportedFuncs = cfg.Generate.TestExportedFuncs
	GenerateTestSubTest = cfg.Generate.TestSubTest
	GenerateTestParallel = cfg.Generate.TestSubTest
	GenerateTestTemplateDir = cfg.Generate.TestTemplateDir
	GenerateTestTemplateParamsPath = cfg.Generate.TemplateParamsPath

	// Guru
	GuruReflection = cfg.Guru.Reflection
	GuruKeepCursor = cfg.Guru.KeepCursor
	GuruJumpFirst = cfg.Guru.JumpFirst

	// Iferr
	IferrAutosave = cfg.Iferr.Autosave

	// Lint
	GolintAutosave = cfg.Lint.GolintAutosave
	GolintIgnore = cfg.Lint.GolintIgnore
	GolintMinConfidence = cfg.Lint.GolintMinConfidence
	GolintMode = cfg.Lint.GolintMode
	GoVetAutosave = cfg.Lint.GoVetAutosave
	GoVetFlags = cfg.Lint.GoVetFlags
	GoVetIgnore = cfg.Lint.GoVetIgnore
	MetalinterAutosave = cfg.Lint.MetalinterAutosave
	MetalinterAutosaveTools = cfg.Lint.MetalinterAutosaveTools
	MetalinterTools = cfg.Lint.MetalinterTools
	MetalinterDeadline = cfg.Lint.MetalinterDeadline
	MetalinterSkipDir = cfg.Lint.MetalinterSkipDir

	// Rename
	RenamePrefill = cfg.Rename.Prefill

	// Terminal
	TerminalMode = cfg.Terminal.Mode
	TerminalPosition = cfg.Terminal.Position
	TerminalHeight = cfg.Terminal.Height
	TerminalWidth = cfg.Terminal.Width
	TerminalStopInsert = cfg.Terminal.StopInsert

	// Test
	TestAutosave = cfg.Test.Autosave
	TestAll = cfg.Test.AllPackage
	TestFlags = cfg.Test.Flags

	// Debug
	DebugEnable = cfg.Debug.Enable
	DebugPprof = cfg.Debug.Pprof
}
