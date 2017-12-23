// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nvim-go/config"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

// CmdVetEval struct type for Eval of GoBuild command.
type CmdVetEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func (c *Command) cmdVet(args []string, eval *CmdVetEval) {
	errch := make(chan interface{}, 1)
	go func() {
		delete(c.buildctxt.Errlist, "Vet") // cleanup
		errch <- c.Vet(args, eval)
	}()

	switch err := <-errch; e := err.(type) {
	case error:
		nvimutil.ErrorWrap(c.Nvim, e)
	case []*nvim.QuickfixError:
		c.buildctxt.Errlist["Vet"] = e
		nvimutil.ErrorList(c.Nvim, c.buildctxt.Errlist, true)
	}
}

// Vet is a simple checker for static errors in Go source code use go tool vet command.
func (c *Command) Vet(args []string, eval *CmdVetEval) interface{} {
	defer nvimutil.Profile(time.Now(), "GoVet")

	vetCmd := exec.Command("go", "tool", "vet")
	vetCmd.Dir = eval.Cwd

	switch {
	case len(args) > 0:
		lastArg := args[len(args)-1]
		if !strings.HasPrefix(lastArg, "-") {
			switch path := filepath.Join(eval.Cwd, lastArg); {
			case args[0] == ".":
				vetCmd.Args = append(vetCmd.Args, ".")
			case pathutil.IsDir(path):
				eval.Cwd = path
				vetCmd.Args = append(vetCmd.Args, args[:len(args)-1]...)
			case pathutil.IsExist(path) && pathutil.IsGoFile(path):
				vetCmd.Args = append(vetCmd.Args, path)
			case filepath.Base(path) == "%":
				path = eval.File
				vetCmd.Args = append(vetCmd.Args, path)
			default:
				err := errors.New("Invalid directory path")
				return errors.WithStack(err)
			}
		} else {
			vetCmd.Args = append(vetCmd.Args, args...)
			vetCmd.Args = append(vetCmd.Args, ".")
		}
	case len(config.GoVetFlags) > 0:
		vetCmd.Args = append(vetCmd.Args, config.GoVetFlags...)
		vetCmd.Args = append(vetCmd.Args, ".")
	default:
		vetCmd.Args = append(vetCmd.Args, ".")
	}

	var stderr bytes.Buffer
	vetCmd.Stderr = &stderr

	vetErr := vetCmd.Run()
	if vetErr != nil {
		errlist, err := nvimutil.ParseError(stderr.Bytes(), eval.Cwd, &c.buildctxt.Build, config.GoVetIgnore)
		if err != nil {
			return errors.WithStack(err)
		}
		return errlist
	}

	return nil
}

func (c *Command) cmdVetComplete(v *nvim.Nvim, a *nvim.CommandCompletionArgs, dir string) ([]string, error) {
	// Flags:
	//  -all
	//        enable all non-experimental checks
	//  -asmdecl
	//        check assembly against Go declarations
	//  -assign
	//        check for useless assignments
	//  -atomic
	//        check for common mistaken usages of the sync/atomic package
	//  -bool
	//        check for mistakes involving boolean operators
	//  -buildtags
	//        check that +build tags are valid
	//  -cgocall
	//        check for types that may not be passed to cgo calls
	//  -composites
	//        check that composite literals used field-keyed elements
	//  -compositewhitelist
	//        use composite white list; for testing only (default true)
	//  -copylocks
	//        check that locks are not passed by value
	//  -lostcancel
	//        check for failure to call cancelation function returned by context.WithCancel
	//  -methods
	//        check that canonically named methods are canonically defined
	//  -nilfunc
	//        check for comparisons between functions and nil
	//  -printf
	//        check printf-like invocations
	//  -printfuncs string
	//        comma-separated list of print function names to check
	//  -rangeloops
	//        check that range loop variables are used correctly
	//  -shadow
	//        check for shadowed variables (experimental; must be set explicitly)
	//  -shadowstrict
	//        whether to be strict about shadowing; can be noisy
	//  -shift
	//        check for useless shifts
	//  -structtags
	//        check that struct field tags have canonical format and apply to exported fields as needed
	//  -tags string
	//        comma-separated list of build tags to apply when parsing
	//  -tests
	//        check for common mistaken usages of tests/documentation examples
	//  -unreachable
	//        check for unreachable code
	//  -unsafeptr
	//        check for misuse of unsafe.Pointer
	//  -unusedfuncs string
	//        comma-separated list of functions whose results must be used (default "errors.New,fmt.Errorf,fmt.Sprintf,fmt.Sprint,sort.Reverse")
	//  -unusedresult
	//        check for unused result of calls to functions in -unusedfuncs list and methods in -unusedstringmethods list
	//  -unusedstringmethods string
	//        comma-separated list of names of methods of type func() string whose results must be used (default "Error,String")
	//  -v
	//        verbose
	complete, err := nvimutil.CompleteFiles(v, a, dir)
	if err != nil {
		return nil, err
	}

	complete = append(complete, []string{
		"-all",
		"-asmdecl",
		"-assign",
		"-atomic",
		"-bool",
		"-buildtags",
		"-cgocall",
		"-composites",
		"-compositewhitelist",
		"-copylocks",
		"-lostcancel",
		"-methods",
		"-nilfunc",
		"-printf",
		"-printfuncs", // arg: string
		"-rangeloops",
		"-shadow",
		"-shadowstrict",
		"-shift",
		"-structtags",
		"-tags", // arg: string
		"-tests",
		"-unreachable",
		"-unsafeptr",
		"-unusedfuncs", // arg: string
		"-unusedresult",
		"-unusedstringmethods", // arg: string
		"-v",
	}...)
	return complete, nil
}
