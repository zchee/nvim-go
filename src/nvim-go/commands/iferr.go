// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"bytes"
	"go/build"
	"go/format"
	"go/parser"
	"go/types"
	"path/filepath"
	"time"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"

	"github.com/juju/errors"
	"github.com/motemen/go-iferr"
	"github.com/neovim-go/vim"
	"golang.org/x/tools/go/loader"
)

const pkgIferr = "GoIferr"

func cmdIferr(v *vim.Vim, file string) {
	go Iferr(v, file)
}

// Iferr automatically insert 'if err' Go idiom by parse the current buffer's Go abstract syntax tree(AST).
func Iferr(v *vim.Vim, file string) error {
	defer profile.Start(time.Now(), "GoIferr")

	dir := filepath.Dir(file)
	ctxt := new(context.Context)
	defer ctxt.Build.SetContext(dir)()

	b, err := v.CurrentBuffer()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgIferr))
	}

	buflines, err := v.BufferLines(b, 0, -1, true)
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgIferr))
	}

	conf := loader.Config{
		ParserMode:  parser.ParseComments,
		TypeChecker: types.Config{FakeImportC: true, DisableUnusedImportCheck: true},
		Build:       &build.Default,
		Cwd:         dir,
		AllowErrors: true,
	}

	var src bytes.Buffer
	src.Write(nvim.ToByteSlice(buflines))

	f, err := conf.ParseFile(file, src.Bytes())
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgIferr))
	}

	conf.CreateFromFiles(file, f)
	prog, err := conf.Load()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgIferr))
	}

	// Reuse src variable
	src.Reset()

	for _, pkg := range prog.InitialPackages() {
		for _, f := range pkg.Files {
			iferr.RewriteFile(prog.Fset, f, pkg.Info)
			format.Node(&src, prog.Fset, f)
		}
	}

	// format.Node() will added pointless newline
	buf := bytes.TrimSuffix(src.Bytes(), []byte{'\n'})
	return v.SetBufferLines(b, 0, -1, true, nvim.ToBufferLines(buf))
}
