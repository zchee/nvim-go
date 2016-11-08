// Copyright 2016 The nvim-go Authors. All rights reserved.
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
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	vim "github.com/neovim/go-client/nvim"
)

func (c *Commands) cmdMetalinter(cwd string) {
	go c.Metalinter(cwd)
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
func (c *Commands) Metalinter(cwd string) error {
	defer nvimutil.Profile(time.Now(), "GoMetaLinter")
	defer c.ctxt.SetContext(cwd)()

	var (
		loclist []*vim.QuickfixError
		b       vim.Buffer
		w       vim.Window
	)
	if c.p == nil {
		c.p = c.v.NewPipeline()
	}
	c.p.CurrentBuffer(&b)
	c.p.CurrentWindow(&w)
	if err := c.p.Wait(); err != nil {
		return err
	}

	var args []string
	switch c.ctxt.Build.Tool {
	case "go":
		args = append(args, cwd+"/...")
	case "gb":
		args = append(args, c.ctxt.Build.ProjectRoot+"/...")
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
		loclist = append(loclist, &vim.QuickfixError{
			FileName: pathutil.Rel(r.Path, cwd),
			LNum:     r.Line,
			Col:      r.Col,
			Text:     r.Linter + ": " + r.Message,
			Type:     strings.ToUpper(r.Severity[:1]),
		})
	}

	if err := nvimutil.SetLoclist(c.v, loclist); err != nil {
		return nvimutil.Echomsg(c.v, "Gometalinter: %v", err)
	}
	return nvimutil.OpenLoclist(c.v, w, loclist, true)
}

type byPath []metalinterResult

func (a byPath) Len() int           { return len(a) }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool { return a[i].Path < a[j].Path }
