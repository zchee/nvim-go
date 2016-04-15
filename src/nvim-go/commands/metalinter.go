// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/gb"
	"nvim-go/nvim"
)

func init() {
	plugin.HandleCommand("Gometalinter",
		&plugin.CommandOptions{
			Eval: "[getcwd(), g:go#lint#metalinter#tools, g:go#lint#metalinter#deadline]"},
		cmdMetalinter)
}

type CmdMetalinterEval struct {
	Cwd      string `msgpack:",array"`
	Tools    []string
	Deadline string
}

func cmdMetalinter(v *vim.Vim, eval CmdMetalinterEval) {
	go Metalinter(v, eval)
}

type metalinterResult struct {
	Linter   string `json:"linter"`   // name of linter tool
	Severity string `json:"severity"` // result of type
	Path     string `json:"path"`     // path of file
	Line     int    `json:"line"`     // line of file
	Col      int    `json:"col"`      // col of file
	Message  string `json:"message"`  // description of linter message
}

func Metalinter(v *vim.Vim, eval CmdMetalinterEval) error {
	defer gb.WithGoBuildForPath(eval.Cwd)()

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

	args := []string{eval.Cwd + "/...", "--json", "--disable-all", "--deadline", eval.Deadline}
	for _, t := range eval.Tools {
		args = append(args, "--enable", t)
	}

	cmd := exec.Command("gometalinter", args...)
	cmd.Dir = eval.Cwd

	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
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
