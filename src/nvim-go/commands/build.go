// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/gb"
	"nvim-go/nvim"
)

func init() {
	plugin.HandleCommand("Gobuild", &plugin.CommandOptions{Eval: "expand('%:p:h')"}, Build)
	plugin.HandleAutocmd("BufWritePost",
		&plugin.AutocmdOptions{Pattern: "*.go", Eval: "[expand('%:p:h'), g:go#build#autobuild]"},
		autocmdBuild)

	log.Debugln("GoBuild Start")
}

func cmdBuild(v *vim.Vim, dir string) {
	go Build(v, dir)
}

type onBuildEval struct {
	Dir  string `msgpack:",array"`
	Flag int64
}

func autocmdBuild(v *vim.Vim, eval onBuildEval) {
	if eval.Flag != int64(0) {
		go Build(v, eval.Dir)
	}
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
		return err
	}

	tmpdir, err := ioutil.TempDir("", "nvim-go")
	if err != nil {
		nvim.Echomsg(v, "%v", err)
	}
	defer os.RemoveAll(tmpdir)

	var compile_cmd string
	currentGopath := strings.Split(build.Default.GOPATH, ":")[0]
	if currentGopath == os.Getenv("GOPATH") {
		compile_cmd = "go"
	} else {
		compile_cmd = "gb"
	}

	cmd := exec.Command(compile_cmd, "build", ".")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	cmd.Run()

	s, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if s.ExitStatus() > 0 {
		loclist := nvim.ParseError(v, string(out))
		if err := nvim.SetLoclist(p, loclist); err != nil {
			nvim.Echomsg(v, "GoBuild: %s", err)
		}
		return nvim.OpenLoclist(p, w, loclist, true)
	} else {
		nvim.Echohl(v, "GoBuild: ", "Function", "SUCCESS")
	}

	return nil
}
