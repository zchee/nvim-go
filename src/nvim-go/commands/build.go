// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/gb"
	"nvim-go/nvim"
)

func init() {
	plugin.HandleCommand("Gobuild", &plugin.CommandOptions{NArgs: "?", Eval: "expand('%:p:h')"}, Build)

	log.Debugln("GoBuild Start")
}

func cmdBuild(v *vim.Vim, args []string, dir string) {
	go Build(v, args, dir)
}

func Build(v *vim.Vim, args []string, dir string) error {
	defer gb.WithGoBuildForPath(dir)()
	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir("", "nvim-go")
	if err != nil {
		nvim.Echomsg(v, "%v", err)
	}
	defer os.RemoveAll(tmpdir)

	// cmd := exec.Command("gb", "build", "-gcflags", "'-h'", ".")
	cmd := exec.Command("gb", "build", ".")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	cmd.Run()

	s, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if s.ExitStatus() > 0 {
		nvim.Echomsg(v, "%s", out)
	} else {
		nvim.Echohl(v, "GoBuild: ", "Function", "SUCCESS")
	}

	return nil
}
