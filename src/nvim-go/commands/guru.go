// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// guru: a tool for answering questions about Go source code.
//
//    http://golang.org/s/oracle-design
//    http://golang.org/s/oracle-user-manual

package commands

import (
	"bytes"
	"fmt"
	"go/build"
	"go/token"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/internal/guru"
	"nvim-go/internal/guru/serial"
	"nvim-go/nvim"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"
	"nvim-go/pathutil"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
	json "github.com/pquerna/ffjson/ffjson"
	"golang.org/x/tools/go/buildutil"
)

var pkgGuru = "Guru"

func init() {
	plugin.HandleFunction("GoGuru", &plugin.FunctionOptions{Eval: "[getcwd(), expand('%:p'), &modified, line2byte(line('.')) + (col('.')-2)]"}, funcGuru)
}

type funcGuruEval struct {
	Cwd      string `msgpack:",array"`
	File     string
	Modified int
	Offset   int
}

func funcGuru(v *vim.Vim, args []string, eval *funcGuruEval) {
	go Guru(v, args, eval)
}

// Guru go source analysis and output result to the quickfix or locationlist.
func Guru(v *vim.Vim, args []string, eval *funcGuruEval) error {
	defer profile.Start(time.Now(), "Guru")
	mode := args[0]
	if len(args) > 1 {
		return guruHelp(v, mode)
	}

	ctxt := new(context.Build)
	dir, _ := filepath.Split(eval.File)
	defer ctxt.SetContext(dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgGuru))
	}

	var scopeFlag []string
	if mode != "definition" {
		switch ctxt.Tool {
		case "go":
			pkgPath := strings.TrimPrefix(pathutil.PackagePath(dir), "src"+string(filepath.Separator))
			scopeFlag = []string{pkgPath + string(filepath.Separator) + "..."}
		case "gb":
			projectName := pathutil.GbProjectName(dir, ctxt.ProjectDir)
			scopeFlag = append(scopeFlag, projectName+string(filepath.Separator)+"...")
		}
	}

	guruContext := &build.Default

	// https://github.com/golang/tools/blob/master/cmd/guru/main.go
	if eval.Modified != 0 {
		overlay := make(map[string][]byte)

		var (
			buf [][]byte
		)

		p.BufferLines(b, 0, -1, true, &buf)
		if err := p.Wait(); err != nil {
			return nvim.ErrorWrap(v, errors.Annotate(err, pkgGuru))
		}

		overlay[eval.File] = bytes.Join(buf, []byte{'\n'})
		guruContext = buildutil.OverlayContext(guruContext, overlay)
	}

	var outputMu sync.Mutex
	var loclist []*quickfix.ErrorlistData
	var (
		fname     string
		line, col int
		err       error
	)
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		outputMu.Lock()
		defer outputMu.Unlock()
		if loclist, fname, line, col, err = parseResult(mode, fset, qr.JSON(fset), eval.Cwd); err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgGuru))
		}
	}

	query := guru.Query{
		Output:     output,
		Pos:        fmt.Sprintf("%s:#%d", eval.File, eval.Offset),
		Build:      guruContext,
		Scope:      scopeFlag,
		Reflection: config.GuruReflection,
	}

	if mode != "definition" {
		nvim.EchoProgress(v, pkgGuru, "analysing")
	}
	if err := guru.Run(mode, &query); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgGuru))
	}

	if err := quickfix.SetLoclist(v, loclist); err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgGuru))
	}

	// jumpfirst or definition mode
	if config.GuruJumpFirst || mode == "definition" {
		p.Command(fmt.Sprintf("edit %s", fname))
		p.SetWindowCursor(w, [2]int{line, col - 1})
		p.Command(`lclose | normal! zz`)
		return nil
	}

	// not definition mode
	return quickfix.OpenLoclist(v, w, loclist, config.GuruKeepCursor)
}

func parseResult(mode string, fset *token.FileSet, data []byte, cwd string) ([]*quickfix.ErrorlistData, string, int, int, error) {
	var (
		loclist []*quickfix.ErrorlistData
		fname   string
		line    int
		col     int
		text    string
	)

	switch mode {

	case "callees":
		var value = serial.Callees{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range value.Callees {
			fname, line, col = quickfix.SplitPos(v.Pos, cwd)
			text = value.Desc + ": " + v.Name
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callers":
		var value = []serial.Caller{}
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range value {
			fname, line, col = quickfix.SplitPos(v.Pos, cwd)
			text = v.Desc + ": " + v.Caller
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callstack":
		var value = serial.CallStack{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range value.Callers {
			fname, line, col = quickfix.SplitPos(v.Pos, cwd)
			text = v.Desc + " " + value.Target
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "definition":
		var value = serial.Definition{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		fname, line, col = quickfix.SplitPos(value.ObjPos, cwd)
		text = value.Desc
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "describe":
		var value = serial.Describe{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		fname, line, col = quickfix.SplitPos(value.Value.ObjPos, cwd)
		text = value.Desc + " " + value.Value.Type
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "freevars":
		var value = serial.FreeVar{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		fname, line, col = quickfix.SplitPos(value.Pos, cwd)
		text = value.Kind + " " + value.Type + " " + value.Ref
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "implements":
		var value = serial.Implements{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range value.AssignableFrom {
			fname, line, col := quickfix.SplitPos(v.Pos, cwd)
			text = v.Kind + " " + v.Name
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "peers":
		var value = serial.Peers{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		fname, line, col := quickfix.SplitPos(value.Pos, cwd)
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Base: selected channel op (<-)",
		})
		for _, v := range value.Allocs {
			fname, line, col := quickfix.SplitPos(v, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Allocs: make(chan) ops",
			})
		}
		for _, v := range value.Sends {
			fname, line, col := quickfix.SplitPos(v, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Sends: ch<-x ops",
			})
		}
		for _, v := range value.Receives {
			fname, line, col := quickfix.SplitPos(v, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Receives: <-ch ops",
			})
		}
		for _, v := range value.Closes {
			fname, line, col := quickfix.SplitPos(v, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Closes: close(ch) ops",
			})
		}

	case "pointsto":
		var value = []serial.PointsTo{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range value {
			fname, line, col := quickfix.SplitPos(v.NamePos, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "type: " + v.Type,
			})
			if len(v.Labels) > -1 {
				for _, vl := range v.Labels {
					fname, line, col := quickfix.SplitPos(vl.Pos, cwd)
					loclist = append(loclist, &quickfix.ErrorlistData{
						FileName: fname,
						LNum:     line,
						Col:      col,
						Text:     vl.Desc,
					})
				}
			}
		}

	case "referrers":
		var packages = serial.ReferrersPackage{}
		if err := json.UnmarshalFast(data, &packages); err != nil {
			return loclist, "", 0, 0, err
		}
		for _, v := range packages.Refs {
			fname, line, col := quickfix.SplitPos(v.Pos, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     v.Text,
			})
		}

	case "whicherrs":
		var value = serial.WhichErrs{}
		err := json.UnmarshalFast(data, &value)
		if err != nil {
			return loclist, "", 0, 0, err
		}
		fname, line, col := quickfix.SplitPos(value.ErrPos, cwd)
		loclist = append(loclist, &quickfix.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Errror Position",
		})
		for _, vg := range value.Globals {
			fname, line, col := quickfix.SplitPos(vg, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Globals",
			})
		}
		for _, vc := range value.Constants {
			fname, line, col := quickfix.SplitPos(vc, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Constants",
			})
		}
		for _, vt := range value.Types {
			fname, line, col := quickfix.SplitPos(vt.Position, cwd)
			loclist = append(loclist, &quickfix.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Types: " + vt.Type,
			})
		}

	}

	if len(loclist) == 0 {
		return loclist, "", 0, 0, fmt.Errorf("%s not fount", mode)
	}
	return loclist, fname, line, col, nil
}

func guruHelp(v *vim.Vim, mode string) error {
	switch mode {
	case "callees":
		return nvim.EchohlBefore(v, "GoGuruCallees", "Function", "Show possible targets of selected function call")
	case "callers":
		return nvim.EchohlBefore(v, "GoGuruCallers", "Function", "Show possible callers of selected function")
	case "callstack":
		return nvim.EchohlBefore(v, "GoGuruCallstack", "Function", "Show path from callgraph root to selected function")
	case "definition":
		return nvim.EchohlBefore(v, "GoGuruDefinition", "Function", "Show declaration of selected identifier")
	case "describe":
		return nvim.EchohlBefore(v, "GoGuruDescribe", "Function", "Describe selected syntax: definition, methods, etc")
	case "freevars":
		return nvim.EchohlBefore(v, "GoGurufreevars", "Function", "Show free variables of selection")
	case "implements":
		return nvim.EchohlBefore(v, "GoGuruImplements", "Function", "Show 'implements' relation for selected type or method")
	case "peers":
		return nvim.EchohlBefore(v, "GoGuruChannelPeers", "Function", "Show send/receive corresponding to selected channel op")
	case "pointsto":
		return nvim.EchohlBefore(v, "GoGuruPointsto", "Function", "Show variables the selected pointer may point to")
	case "referrers":
		return nvim.EchohlBefore(v, "GoGuruReferrers", "Function", "Show all refs to entity denoted by selected identifier")
	case "what":
		return nvim.EchohlBefore(v, "GoGuruWhat", "Function", "Show basic information about the selected syntax node")
	case "whicherrs":
		return nvim.EchohlBefore(v, "GoGuruWhicherrs", "Function", "Show possible values of the selected error variable")
	}

	return nvim.Echoerr(v, "Invalid arguments")
}
