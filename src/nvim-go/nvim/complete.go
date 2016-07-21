// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"nvim-go/pathutil"

	"github.com/neovim-go/vim"
)

// CompleteFiles provides a "-complete=file" completion exclude the non go files.
func CompleteFiles(v *vim.Vim, a *vim.CommandCompletionArgs, dir string) (filelist []string, err error) {
	switch {
	case len(a.ArgLead) > 0:
		a.ArgLead = filepath.Clean(a.ArgLead)
		if pathutil.IsDir(a.ArgLead) { // abs or rel directory path
			files, err := ioutil.ReadDir(a.ArgLead)
			if err != nil {
				return nil, err
			}
			for _, f := range files {
				if filepath.Ext(f.Name()) == ".go" || f.IsDir() {
					filelist = append(filelist, a.ArgLead+string(filepath.Separator)+f.Name())
				}
			}
		} else { // lacking directory path or filename
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				return nil, err
			}
			return matchFile(files, a.ArgLead), nil
		}

	default:
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".go" || f.IsDir() {
				filelist = append(filelist, f.Name())
			}
		}
	}

	return filelist, nil
}

func matchFile(files []os.FileInfo, filename string) []string {
	var filelist []string
	for _, f := range files {
		if f.Name() == filename || strings.HasPrefix(f.Name(), filename) {
			filelist = append(filelist, f.Name())
		}
	}
	return filelist
}
