// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"go/build"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/gb"
	"nvim-go/nvim"
	"nvim-go/pkgs"
)

func init() {
	plugin.HandleCommand("Gobuild", &plugin.CommandOptions{Eval: "expand('%:p:h')"}, Build)
}

func cmdBuild(v *vim.Vim, dir string) {
	go Build(v, dir)
}

func Build(v *vim.Vim, dir string) error {
	defer gb.WithGoBuildForPath(dir)()
	var (
		b vim.Buffer
		w vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return nvim.Echoerr(v, err)
	}

	var compile_cmd string
	currentGoPath := strings.Split(build.Default.GOPATH, ":")[0]
	if currentGoPath == os.Getenv("GOPATH") {
		compile_cmd = "go"
	} else {
		compile_cmd = "gb"
	}

	rootDir := pkgs.FindvcsDir(dir)

	cmd := exec.Command(compile_cmd, "build")
	cmd.Dir = rootDir
	out, _ := cmd.CombinedOutput()
	cmd.Run()

	s, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if s.ExitStatus() > 0 {
		loclist := nvim.ParseError(v, string(out), dir)
		if err := nvim.SetLoclist(p, loclist); err != nil {
			return nvim.Echoerr(v, err)
		}
		return nvim.OpenLoclist(p, w, loclist, true)
	}

	return nvim.Echohl(v, "GoBuild: ", "Function", "SUCCESS")
}
