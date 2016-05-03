// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/nvim"

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
	defer nvim.Profile(time.Now(), "GoMetaLinter")
	defer context.SetContext(cwd)()

	var (
		loclist []*nvim.ErrorlistData
		b       vim.Buffer
		w       vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	args := []string{cwd + "/...", "--json", "--disable-all", "--deadline", config.MetalinterDeadline}
	for _, t := range config.MetalinterTools {
		args = append(args, "--enable", t)
	}
	if len(config.MetalinterSkipDir) != 0 {
		for _, dir := range config.MetalinterSkipDir {
			args = append(args, "--skip", dir)
		}
	}

	cmd := exec.Command("gometalinter", args...)
	cmd.Dir = cwd
	stdout, _ := cmd.Output()
	cmd.Run()

	var result = []metalinterResult{}
	if err := json.Unmarshal(stdout, &result); err != nil {
		fmt.Println(err)
	}

	for _, r := range result {
		var errorType string
		switch r.Severity {
		case "error":
			errorType = "E"
		case "warning":
			errorType = "W"
		}
		loclist = append(loclist, &nvim.ErrorlistData{
			FileName: r.Path,
			LNum:     r.Line,
			Col:      r.Col,
			Text:     r.Linter + ": " + r.Message,
			Type:     errorType,
		})
	}

	if err := nvim.SetLoclist(p, loclist); err != nil {
		return nvim.Echomsg(v, "Gometalinter: %v", err)
	}
	return nvim.OpenLoclist(p, w, loclist, true)
}
