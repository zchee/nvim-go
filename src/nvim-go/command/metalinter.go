// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/pathutil"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

func (c *Command) cmdMetalinter(cwd string) {
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
func (c *Command) Metalinter(cwd string) error {
	defer nvimutil.Profile(time.Now(), "GoMetaLinter")

	var loclist []*nvim.QuickfixError
	w := nvim.Window(c.ctx.WinID)

	var args []string
	switch c.ctx.Build.Tool {
	case "go":
		args = append(args, cwd+"/...")
	case "gb":
		args = append(args, c.ctx.Build.ProjectRoot+"/...")
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
			return errors.WithStack(err)
		}
	}

	sort.Sort(byPath(result))

	for _, r := range result {
		loclist = append(loclist, &nvim.QuickfixError{
			FileName: pathutil.Rel(r.Path, cwd),
			LNum:     r.Line,
			Col:      r.Col,
			Text:     r.Linter + ": " + r.Message,
			Type:     strings.ToUpper(r.Severity[:1]),
		})
	}

	if err := nvimutil.SetLoclist(c.Nvim, loclist); err != nil {
		return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
	}
	return nvimutil.OpenLoclist(c.Nvim, w, loclist, true)
}

type byPath []metalinterResult

func (a byPath) Len() int           { return len(a) }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool { return a[i].Path < a[j].Path }
