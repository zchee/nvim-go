// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/haya14busa/errorformat"
	"github.com/neovim/go-client/nvim"
	"go.opencensus.io/trace"
	"go.uber.org/zap"

	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/logger"
)

// regexp pattern: https://regex101.com/r/bUVZpH/2
var errRe = regexp.MustCompile(`(?m)^(?:#\s([[:graph:]]+))?(?:[\s\t]+)?([^\s:]+):(\d+)(?::(\d+))?(?::)?\s(.*)`)

// ParseError parses a typical Go tools error messages.
func ParseError(ctx context.Context, errmsg []byte, cwd string, bctxt *buildctxt.Build, ignoreDirs []string) ([]*nvim.QuickfixError, error) {
	defer Profile(ctx, time.Now(), "ParseError")
	span := trace.FromContext(ctx)
	span.SetName("ParseError")
	defer span.End()

	if config.IsDebug() {
		log := logger.FromContext(ctx)
		efm, err := errorformat.NewErrorformat([]string{`%f:%l:%c: %m`, `%-G%.%#`})
		if err != nil {
			return nil, err
		}
		s := efm.NewScanner(bytes.NewReader(errmsg))
		for s.Scan() {
			log.Debug("errorformat: `%%f:%l:%%c: %m`, `%%-G%.%%#`", zap.Any("s.Entry", s.Entry()))
		}
	}

	var (
		// packagePath for the save the error files parent directory.
		// It will be re-assigned if "# " is in the error message.
		packagePath string
		errlist     []*nvim.QuickfixError
	)

	// m[1]: package path with "# " prefix
	// m[2]: error files relative path
	// m[3]: line number of error point
	// m[4]: column number of error point
	// m[5]: error description text
	for _, m := range errRe.FindAllSubmatch(errmsg, -1) {
		if m[1] != nil {
			// Save the package path for the second subsequent errors
			packagePath = string(m[1])
		}
		filename := string(m[2])

		// Avoid the local package error. like "package foo" and edit "cmd/foo/main.go"
		if !filepath.IsAbs(filename) && packagePath != "" {
			// Joins the packagePath and error file
			filename = filepath.Join(packagePath, filepath.Base(filename))
		}

		// Cleanup filename to relative path of current working directory
		switch bctxt.Tool {
		case "go":
			var sep string
			switch {
			// filename has not directory path
			case filepath.Dir(filename) == ".":
				filename = filepath.Join(cwd, filename)
			// not contains '#' package title in errror
			case strings.HasPrefix(filename, cwd):
				sep = cwd
				filename = strings.TrimPrefix(filename, sep+string(filepath.Separator))
			// filename is like "github.com/foo/bar.go"
			case strings.HasPrefix(filename, fs.TrimGoPath(cwd)):
				sep = fs.TrimGoPath(cwd) + string(filepath.Separator)
				filename = strings.TrimPrefix(filename, sep)
			default:
				filename = fs.JoinGoPath(filename)
			}
		case "gb":
			// gb compiler error messages is relative filename path of project root dir
			if !filepath.IsAbs(filename) {
				filename = filepath.Join(bctxt.ProjectRoot, "src", filename)
			}
		default:
			return nil, errors.New("unknown compiler tool")
		}

		// Finally, try to convert the relative path from cwd
		filename = fs.Rel(cwd, filename)
		if ignoreDirs != nil {
			if contains(filename, ignoreDirs) {
				continue
			}
		}

		// line is necessary for error messages
		line, err := strconv.Atoi(string(m[3]))
		if err != nil {
			return nil, err
		}

		// Ignore err because fail strconv.Atoi will assign 0 to col
		col, _ := strconv.Atoi(string(m[4]))

		errlist = append(errlist, &nvim.QuickfixError{
			FileName: filename,
			LNum:     line,
			Col:      col,
			Text:     string(bytes.TrimSpace(m[5])),
		})
	}

	return errlist, nil
}

func contains(s string, substr []string) bool {
	for _, str := range substr {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}
