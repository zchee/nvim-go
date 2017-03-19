// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// guru: a tool for answering questions about Go source code.
//
//    http://golang.org/s/using-guru

package commands

import (
	"bytes"
	"fmt"
	"go/build"
	"go/token"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nvim-go/config"
	"nvim-go/internal/guru"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	"github.com/davecgh/go-spew/spew"
	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"golang.org/x/tools/cmd/guru/serial"
	"golang.org/x/tools/go/buildutil"
)

type funcGuruEval struct {
	Cwd      string `msgpack:",array"`
	File     string
	Modified int
	Offset   int
}

func (c *Commands) funcGuru(args []string, eval *funcGuruEval) {
	go func() {
		err := c.Guru(args, eval)

		switch err := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, err)
		default:
			// nothing to do
		}
	}()
}

// Guru go source analysis and output result to the quickfix or locationlist.
func (c *Commands) Guru(args []string, eval *funcGuruEval) interface{} {
	defer nvimutil.Profile(time.Now(), "Guru")

	mode := args[0]
	if len(args) > 1 {
		return guruHelp(c.Nvim, mode)
	}

	defer func() (err error) {
		if r := recover(); r != nil {
			err = errors.Errorf("guru internal panic.\nMaybe your set 'g:go#guru#reflection' to 1. Please retry with disable it option.\nOriginal panic message:\n\t%v", r.(error))
			return errors.WithStack(err)
		}
		return nil
	}()

	b := nvim.Buffer(c.ctx.BufNr)
	w := nvim.Window(c.ctx.WinID)
	batch := c.Nvim.NewBatch()

	guruContext := &build.Default

	// https://github.com/golang/tools/blob/master/cmd/guru/main.go
	if eval.Modified != 0 {
		overlay := make(map[string][]byte)
		var buf [][]byte

		batch.BufferLines(b, 0, -1, true, &buf)
		if err := batch.Execute(); err != nil {
			return errors.WithStack(err)
		}

		overlay[eval.File] = bytes.Join(buf, []byte{'\n'})
		guruContext = buildutil.OverlayContext(guruContext, overlay)
	}

	var loclist []*nvim.QuickfixError
	query := guru.Query{
		Pos:        fmt.Sprintf("%s:#%d", eval.File, eval.Offset),
		Build:      guruContext,
		Reflection: config.GuruReflection,
	}

	if mode == "definition" {
		obj, err := Definition(&query)
		if err != nil {
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}
		fname, line, col := nvimutil.SplitPos(obj.ObjPos, eval.Cwd)

		batch.Command("normal! m'")
		// TODO(zchee): should change nvimutil.SplitPos behavior
		f := strings.Split(obj.ObjPos, ":")
		if f[0] != eval.File {
			batch.Command(fmt.Sprintf("keepjumps edit %s", pathutil.Rel(eval.Cwd, fname)))
		}
		batch.SetWindowCursor(w, [2]int{line, col - 1})
		if err := batch.Execute(); err != nil {
			return nvimutil.ErrorWrap(c.Nvim, errors.WithStack(err))
		}

		return c.Nvim.Command(`lclose | normal! zz`)
	}

	var scope string
	switch c.ctx.Build.Tool {
	case "go":
		pkgID, err := pathutil.PackageID(filepath.Dir(eval.File))
		if err != nil {
			return errors.WithStack(err)
		}
		scope = pkgID
	case "gb":
		scope = pathutil.GbProjectName(c.ctx.Build.ProjectRoot)
	}
	query.Scope = append(query.Scope, filepath.Join(scope, "..."))

	var (
		outputMu sync.Mutex
		err      error
	)
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		var err error
		outputMu.Lock()
		defer outputMu.Unlock()

		res := qr.Result(fset)
		if loclist, err = parseResult(mode, res, eval.Cwd); err != nil {
			err = errors.WithStack(err)
		}
	}
	if err != nil {
		return errors.WithStack(err)
	}
	query.Output = output

	nvimutil.EchoProgress(c.Nvim, "Guru", fmt.Sprintf("analysing %s", mode))
	if err := guru.Run(mode, &query); err != nil {
		return errors.WithStack(err)
	}
	if len(loclist) == 0 {
		return fmt.Errorf("%s not fount", mode)
	}

	defer nvimutil.ClearMsg(c.Nvim)
	if err := nvimutil.SetLoclist(c.Nvim, loclist); err != nil {
		return errors.WithStack(err)
	}

	// jumpfirst or definition mode
	if config.GuruJumpFirst {
		batch.Command(`silent ll 1`)
		batch.Command(`normal! zz`)
		return batch.Execute()
	}

	var keepCursor bool
	if int64(1) == config.GuruKeepCursor[mode] {
		keepCursor = true
	}
	return nvimutil.OpenLoclist(c.Nvim, w, loclist, keepCursor)
}

var errTypeAssertion = errors.New("type assertion error")

func parseResult(mode string, res interface{}, cwd string) ([]*nvim.QuickfixError, error) {
	if config.DebugEnable {
		log.Printf("res:\n%+v\n", spew.Sdump(res))
	}
	var (
		loclist []*nvim.QuickfixError
		fname   string
		line    int
		col     int
		text    string
	)

	switch mode {
	case "callees":
		value, ok := res.(*serial.Callees)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, v := range value.Callees {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = value.Desc + ": " + v.Name
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callers":
		value, ok := res.([]serial.Caller)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, v := range value {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = v.Desc + ": " + v.Caller
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callstack":
		value, ok := res.(*serial.CallStack)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, v := range value.Callers {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = v.Desc + " " + value.Target
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "describe":
		value, ok := res.(serial.Describe)
		if !ok {
			return loclist, errTypeAssertion
		}
		fname, line, col = nvimutil.SplitPos(value.Value.ObjPos, cwd)
		text = value.Desc + " " + value.Value.Type
		loclist = append(loclist, &nvim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "freevars":
		value, ok := res.([]serial.FreeVar)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, v := range value {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = v.Kind + " " + v.Type + " " + v.Ref
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "implements":
		value, ok := res.(*serial.Implements)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, values := range [][]serial.ImplementsType{value.AssignableTo, value.AssignableFromPtr, value.AssignableFrom} {
			for _, value := range values {
				fname, line, col := nvimutil.SplitPos(value.Pos, cwd)
				text = value.Kind + " " + value.Name
				loclist = append(loclist, &nvim.QuickfixError{
					FileName: fname,
					LNum:     line,
					Col:      col,
					Text:     text,
				})
			}
		}

	case "peers":
		value, ok := res.(*serial.Peers)
		if !ok {
			return loclist, errTypeAssertion
		}
		fname, line, col := nvimutil.SplitPos(value.Pos, cwd)
		loclist = append(loclist, &nvim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Base: selected channel op (<-)",
		})
		for _, v := range value.Allocs {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Allocs: make(chan) ops",
			})
		}
		for _, v := range value.Sends {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Sends: ch<-x ops",
			})
		}
		for _, v := range value.Receives {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Receives: <-ch ops",
			})
		}
		for _, v := range value.Closes {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Closes: close(ch) ops",
			})
		}

	case "pointsto":
		value, ok := res.([]serial.PointsTo)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, v := range value {
			fname, line, col := nvimutil.SplitPos(v.NamePos, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "type: " + v.Type,
			})
			if len(v.Labels) > -1 {
				for _, vl := range v.Labels {
					fname, line, col := nvimutil.SplitPos(vl.Pos, cwd)
					loclist = append(loclist, &nvim.QuickfixError{
						FileName: fname,
						LNum:     line,
						Col:      col,
						Text:     vl.Desc,
					})
				}
			}
		}

	// TODO(zchee): Support serial.ReferrersInitial type
	case "referrers":
		switch value := res.(type) {
		case serial.ReferrersPackage:
			for _, v := range value.Refs {
				fname, line, col := nvimutil.SplitPos(v.Pos, cwd)
				loclist = append(loclist, &nvim.QuickfixError{
					FileName: fname,
					LNum:     line,
					Col:      col,
					Text:     v.Text,
				})
			}
		default:
			return loclist, errTypeAssertion
		}

	// TODO(zchee): implements what mode

	case "whicherrs":
		value, ok := res.(*serial.WhichErrs)
		if !ok {
			return loclist, errTypeAssertion
		}
		fname, line, col := nvimutil.SplitPos(value.ErrPos, cwd)
		loclist = append(loclist, &nvim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Errror Position",
		})
		for _, vg := range value.Globals {
			fname, line, col := nvimutil.SplitPos(vg, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Globals",
			})
		}
		for _, vc := range value.Constants {
			fname, line, col := nvimutil.SplitPos(vc, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Constants",
			})
		}
		for _, vt := range value.Types {
			fname, line, col := nvimutil.SplitPos(vt.Position, cwd)
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Types: " + vt.Type,
			})
		}

	}
	return loclist, nil
}

func guruHelp(v *nvim.Nvim, mode string) error {
	switch mode {
	case "callees":
		return nvimutil.EchohlBefore(v, "GoGuruCallees", "Function", "Show possible targets of selected function call")
	case "callers":
		return nvimutil.EchohlBefore(v, "GoGuruCallers", "Function", "Show possible callers of selected function")
	case "callstack":
		return nvimutil.EchohlBefore(v, "GoGuruCallstack", "Function", "Show path from callgraph root to selected function")
	case "definition":
		return nvimutil.EchohlBefore(v, "GoGuruDefinition", "Function", "Show declaration of selected identifier")
	case "describe":
		return nvimutil.EchohlBefore(v, "GoGuruDescribe", "Function", "Describe selected syntax: definition, methods, etc")
	case "freevars":
		return nvimutil.EchohlBefore(v, "GoGurufreevars", "Function", "Show free variables of selection")
	case "implements":
		return nvimutil.EchohlBefore(v, "GoGuruImplements", "Function", "Show 'implements' relation for selected type or method")
	case "peers":
		return nvimutil.EchohlBefore(v, "GoGuruChannelPeers", "Function", "Show send/receive corresponding to selected channel op")
	case "pointsto":
		return nvimutil.EchohlBefore(v, "GoGuruPointsto", "Function", "Show variables the selected pointer may point to")
	case "referrers":
		return nvimutil.EchohlBefore(v, "GoGuruReferrers", "Function", "Show all refs to entity denoted by selected identifier")
	case "what":
		return nvimutil.EchohlBefore(v, "GoGuruWhat", "Function", "Show basic information about the selected syntax node")
	case "whicherrs":
		return nvimutil.EchohlBefore(v, "GoGuruWhicherrs", "Function", "Show possible values of the selected error variable")
	default:
		return nvimutil.Echoerr(v, "Invalid arguments")
	}
}
