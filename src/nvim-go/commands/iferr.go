// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/motemen/go-iferr"
	"golang.org/x/tools/go/loader"
)

func init() {
	plugin.HandleCommand("GoIferr", &plugin.CommandOptions{Eval: "expand('%:p')"}, cmdIferr)
}

func cmdIferr(v *vim.Vim, file string) {
	go Iferr(v, file)
}

// Iferr automatically insert 'if err' Go idiom by parse the current buffer's Go abstract syntax tree(AST).
func Iferr(v *vim.Vim, file string) error {
	defer profile.Start(time.Now(), "GoIferr")
	var ctxt = context.Build{}
	dir, _ := filepath.Split(file)
	defer ctxt.SetContext(dir)()

	b, err := v.CurrentBuffer()
	if err != nil {
		return err
	}

	bufline, err := v.BufferLines(b, 0, -1, true)
	if err != nil {
		return err
	}

	var buf string
	for _, bufstr := range bufline {
		buf += "\n" + string(bufstr)
	}

	conf := loader.Config{
		ParserMode:  parser.ParseComments,
		TypeChecker: types.Config{FakeImportC: true, DisableUnusedImportCheck: false},
		Build:       &ctxt.BuildContext,
		Cwd:         dir,
		AllowErrors: true,
	}

	f, err := conf.ParseFile(file, buf)
	if err != nil {
		return nvim.Echoerr(v, "GoIferr: %v", err)
	}

	conf.CreateFromFiles(file, f)
	prog, err := conf.Load()
	if err != nil {
		return err
	}

	saveStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	for _, pkg := range prog.InitialPackages() {
		for _, f := range pkg.Files {
			iferr.RewriteFile(prog.Fset, f, pkg.Info)
			format.Node(w, prog.Fset, f)
		}
	}

	w.Close()
	os.Stdout = saveStdout

	out, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return v.SetBufferLines(b, 0, -1, true, bytes.Split(out, []byte{'\n'}))
}
