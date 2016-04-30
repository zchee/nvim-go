// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("Gobuild", &plugin.CommandOptions{Eval: "getcwd()"}, cmdBuild)
}

func cmdBuild(v *vim.Vim, cwd string) {
	go Build(v, cwd)
}

// Build building the current buffer's package use compile tool that determined from the directory structure.
func Build(v *vim.Vim, cwd string) error {
	defer context.WithGoBuildForPath(cwd)()
	defer nvim.Profile(time.Now(), "GoBuild")

	var (
		b vim.Buffer
		w vim.Window
	)

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return nvim.Echoerr(v, "GoBuild: %v", err)
	}

	baseDir := context.FindVcsRoot(cwd)

	var compiler string

	buildDir := strings.Split(build.Default.GOPATH, string(filepath.ListSeparator))[0]
	gopath := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))[0]
	if buildDir == gopath {
		compiler = "go"
	} else {
		compiler = "gb"
		baseDir = filepath.Join(baseDir, "src")
	}

	tmpfile, err := ioutil.TempFile(os.TempDir(), "nvim-go")
	if err != nil {
		return err
	}

	cmd := exec.Command(compiler, "build", "-o", tmpfile.Name())
	out, _ := cmd.CombinedOutput()

	cmd.Run()

	s, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if s.ExitStatus() > 0 {
		loclist := nvim.ParseError(v, string(out), cwd, baseDir)
		if err := nvim.SetLoclist(p, loclist); err != nil {
			return nvim.Echoerr(v, "GoBuild: %v", err)
		}
		return nvim.OpenLoclist(p, w, loclist, true)
	}

	return nvim.Echohl(v, "GoBuild: ", "Function", "SUCCESS")
}
