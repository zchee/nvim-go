// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gometalinter", &plugin.CommandOptions{Eval: "getcwd()"}, cmdMetalinter)
}

func cmdMetalinter(v *vim.Vim, cwd string) {
	go Metalinter(v, cwd)
}

type metalinterResult struct {
	Linter   string `json:"linter"`   // name of linter tool
	Severity string `json:"severity"` // result of type
	Path     string `json:"path"`     // path of file
	Line     int    `json:"line"`     // line of file
	Col      int    `json:"col"`      // col of file
	Message  string `json:"message"`  // description of linter message
}

// Metalinter lint the Go sources from current buffer's package use gometalinter tool.
func Metalinter(v *vim.Vim, cwd string) error {
	defer profile.Start(time.Now(), "GoMetaLinter")
	var ctxt = context.Build{}
	defer ctxt.SetContext(cwd)()

	var (
		loclist []*quickfix.ErrorlistData
		b       vim.Buffer
		w       vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	var args []string
	switch ctxt.Tool {
	case "go":
		args = append(args, cwd+"/...")
	case "gb":
		args = append(args, ctxt.ProjectDir+"/...")
	}
	args = append(args, []string{"--json", "--disable-all", "--deadline", config.MetalinterDeadline}...)

	for _, t := range config.MetalinterTools {
		args = append(args, "--enable", t)
	}
	if len(config.MetalinterSkipDir) != 0 {
		for _, dir := range config.MetalinterSkipDir {
			args = append(args, "--skip", dir)
		}
	}

	cmd := exec.Command("gometalinter", args...)
	stdout, err := cmd.Output()
	cmd.Run()

	var result = []metalinterResult{}
	if err != nil {
		if err := json.Unmarshal(stdout, &result); err != nil {
			return err
		}
	}

	sort.Sort(byPath(result))

	for _, r := range result {
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: nvim.ToRelPath(r.Path, cwd),
			LNum:     r.Line,
			Col:      r.Col,
			Text:     r.Linter + ": " + r.Message,
			Type:     strings.ToUpper(r.Severity[:1]),
		})
	}

	if err := quickfix.SetLoclist(v, loclist); err != nil {
		return nvim.Echomsg(v, "Gometalinter: %v", err)
	}
	return quickfix.OpenLoclist(v, w, loclist, true)
}

type byPath []metalinterResult

func (a byPath) Len() int           { return len(a) }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool { return a[i].Path < a[j].Path }
