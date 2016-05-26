// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/cweill/gotests/gotests/process"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("GoGenerateTest", &plugin.CommandOptions{NArgs: "*", Complete: "file", Eval: "expand('%:p:h')"}, cmdGenerateTest)
}

func cmdGenerateTest(v *vim.Vim, files []string, dir string) {
	go GenerateTest(v, files, dir)
}

// GenerateTest generates the test files based by current buffer or args files
// functions.
// TODO(zchee): Currently Support '-all' flag only.
// Needs support -excl, -exported, -i, -only flags.
func GenerateTest(v *vim.Vim, files []string, dir string) error {
	defer profile.Start(time.Now(), "GenerateTest")
	var ctxt = context.Build{}
	defer ctxt.SetContext(filepath.Dir(dir))()

	b, err := v.CurrentBuffer()
	if err != nil {
		return nvim.Echoerr(v, "GoGenerateTest: %v", err)
	}

	if len(files) == 0 {
		f, err := v.BufferName(b)
		if err != nil {
			return nvim.Echoerr(v, "GoGenerateTest: %v", err)
		}
		files = append(files, f)
	}

	var opt = process.Options{
		AllFuncs:    true,
		ExclFuncs:   "init",
		WriteOutput: true,
		PrintInputs: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	process.Run(w, files, &opt)

	w.Close()
	os.Stdout = oldStdout

	var genFuncs string
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		genFuncs += scan.Text() + "\n"
	}

	// TODO(zchee): More beautiful code
	suffix := "_test.go "
	var ftests, ftestsRel string
	for _, f := range files {
		fnAbs := strings.Split(f, filepath.Ext(f))
		ftests += fnAbs[0] + suffix

		_, fnRel := filepath.Split(fnAbs[0])
		ftestsRel += fnRel + suffix
	}
	log.Println(ftests, ftestsRel)

	ask := fmt.Sprintf("%s\nGoGenerateTest: Generated %s\nGoGenerateTest: Open it? (y, n): ", genFuncs, ftestsRel)
	var answer interface{}
	if err := v.Call("input", &answer, ask); err != nil {
		return err
	}

	// TODO(zchee): Support open the ftests[0] file only.
	// If passes multiple files for 'edit' commands, occur 'E172: Only one file name allowed' errror.
	if answer.(string) != "n" {
		return v.Command(fmt.Sprintf("edit %s", ftests))
	}

	return nil
}
