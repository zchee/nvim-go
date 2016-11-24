// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"

	"github.com/cweill/gotests/gotests/process"
	"github.com/pkg/errors"
)

func (c *Commands) cmdGenerateTest(args []string, ranges [2]int, bang bool, dir string) {
	go c.GenerateTest(args, ranges, bang, dir)
}

// GenerateTest generates the test files based by current buffer or args files
// functions.
// TODO(zchee): Currently Support '-all' flag only.
// Needs support -exported, -i, -only flags.
func (c *Commands) GenerateTest(args []string, ranges [2]int, bang bool, dir string) error {
	defer nvimutil.Profile(time.Now(), "GenerateTest")
	defer c.ctxt.SetContext(filepath.Dir(dir))()

	b, err := c.Nvim.CurrentBuffer()
	if err != nil {
		return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
	}

	if len(args) == 0 {
		f, err := c.Nvim.BufferName(b)
		if err != nil {
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}
		args = []string{f}
	}

	var opt = process.Options{
		AllFuncs:    true,
		ExclFuncs:   config.GenerateTestExclFuncs,
		WriteOutput: true,
		PrintInputs: true,
		Subtests:    true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	process.Run(w, args, &opt)

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
	for _, f := range args {
		fnAbs := strings.Split(f, filepath.Ext(f))
		ftests += fnAbs[0] + suffix

		_, fnRel := filepath.Split(fnAbs[0])
		ftestsRel += fnRel + suffix
	}

	ask := fmt.Sprintf("%s\nGoGenerateTest: Generated %s\nGoGenerateTest: Open it? (y, n): ", genFuncs, ftestsRel)
	var answer interface{}
	if err := c.Nvim.Call("input", &answer, ask); err != nil {
		return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
	}

	// TODO(zchee): Support open the ftests[0] file only.
	// If passes multiple files for 'edit' commands, occur 'E172: Only one file name allowed' errror.
	if answer.(string) != "n" {
		return c.Nvim.Command(fmt.Sprintf("edit %s", ftests))
	}

	return nil
}
