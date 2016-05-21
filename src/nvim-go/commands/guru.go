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
	"encoding/json"
	"fmt"
	"go/build"
	"go/token"
	"strconv"
	"strings"
	"sync"
	"time"

	"nvim-go/config"
	"nvim-go/context"
	"nvim-go/internal/guru"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"
	"nvim-go/nvim/profile"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/cmd/guru/serial"
	"golang.org/x/tools/go/buildutil"
)

func init() {
	plugin.HandleFunction("GoGuru", &plugin.FunctionOptions{Eval: "[getcwd(), expand('%:p:h'), expand('%:p'), &modified]"}, funcGuru)
}

type funcGuruEval struct {
	Cwd      string `msgpack:",array"`
	Dir      string
	File     string
	Modified int
}

func funcGuru(v *vim.Vim, args []string, eval *funcGuruEval) {
	go Guru(v, args, eval)
}

// Guru go source analysis and output result to the quickfix or locationlist.
func Guru(v *vim.Vim, args []string, eval *funcGuruEval) error {
	defer profile.Start(time.Now(), "Guru")

	mode := args[0]
	if len(args) > 1 {
		arg := args[1]
		if arg == "-h" || arg == "--help" {
			switch mode {
			case "callees":
				return nvim.EchohlBefore(v, "GoGuruCallees", "Function", "show possible targets of selected function call")
			case "callers":
				return nvim.EchohlBefore(v, "GoGuruCallers", "Function", "show possible callers of selected function")
			case "callstack":
				return nvim.EchohlBefore(v, "GoGuruCallstack", "Function", "show path from callgraph root to selected function")
			case "definition":
				return nvim.EchohlBefore(v, "GoGuruDefinition", "Function", "show declaration of selected identifier")
			case "describe":
				return nvim.EchohlBefore(v, "GoGuruDescribe", "Function", "describe selected syntax: definition, methods, etc")
			case "freevars":
				return nvim.EchohlBefore(v, "GoGurufreevars", "Function", "show free variables of selection")
			case "implements":
				return nvim.EchohlBefore(v, "GoGuruImplements", "Function", "show 'implements' relation for selected type or method")
			case "peers":
				return nvim.EchohlBefore(v, "GoGuruChannelPeers", "Function", "show send/receive corresponding to selected channel op")
			case "pointsto":
				return nvim.EchohlBefore(v, "GoGuruPointsto", "Function", "show variables the selected pointer may point to")
			case "referrers":
				return nvim.EchohlBefore(v, "GoGuruReferrers", "Function", "show all refs to entity denoted by selected identifier")
			case "what":
				return nvim.EchohlBefore(v, "GoGuruWhat", "Function", "show basic information about the selected syntax node")
			case "whicherrs":
				return nvim.EchohlBefore(v, "GoGuruWhicherrs", "Function", "show possible values of the selected error variable")
			}
		} else {
			return nvim.Echoerr(v, "Invalid arguments")
		}
	}

	var c = context.Build{}
	defer c.SetContext(eval.Dir)()

	var (
		b vim.Buffer
		w vim.Window
	)
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	dir := strings.Split(eval.Dir, "src/")
	scopeFlag := dir[len(dir)-1]

	pos, err := buffer.ByteOffsetPipe(p, b, w)
	if err != nil {
		return nvim.Echomsg(v, err)
	}

	ctxt := &build.Default

	// https://github.com/golang/tools/blob/master/cmd/guru/main.go
	if eval.Modified != 0 {
		overlay := make(map[string][]byte)

		var (
			buffer [][]byte
			bname  string
		)

		p.BufferLines(b, 0, -1, true, &buffer)
		p.BufferName(b, &bname)
		if err := p.Wait(); err != nil {
			return err
		}

		overlay[bname] = bytes.Join(buffer, []byte{'\n'})
		ctxt = buildutil.OverlayContext(ctxt, overlay)
	}

	var outputMu sync.Mutex
	var loclist []*quickfix.ErrorlistData
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		outputMu.Lock()
		defer outputMu.Unlock()
		if loclist, err = parseResult(mode, fset, qr.JSON(fset), eval.Cwd); err != nil {
			nvim.Echoerr(v, "GoGuru: %v", err)
		}
	}

	query := guru.Query{
		Output:     output,
		Pos:        eval.File + ":#" + strconv.FormatInt(int64(pos), 10),
		Build:      ctxt,
		Scope:      []string{scopeFlag},
		Reflection: config.GuruReflection,
	}

	if err := guru.Run(mode, &query); err != nil {
		return nvim.Echomsg(v, "GoGuru:", err)
	}

	if err := quickfix.SetLoclist(v, loclist); err != nil {
		return nvim.Echomsg(v, "GoGuru:", err)
	}

	// jumpfirst or definition mode
	if config.GuruJumpFirst || mode == "definition" {
		p.Command("silent ll 1 | normal zz")
		// Define the mapping to add 'zz' to <C-o> in the buffer local.
		p.Command("nnoremap <silent><buffer> <C-o> <C-o>zz")
	}

	// not definition mode
	if mode != "definition" {
		return quickfix.OpenLoclist(v, w, loclist, config.GuruKeepCursor)
	}

	return nil
}

func parseResult(mode string, fset *token.FileSet, data []byte, cwd string) ([]*quickfix.ErrorlistData, error) {
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		if err := json.Unmarshal(data, &packages); err != nil {
			return loclist, err
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
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
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
		return loclist, fmt.Errorf("%s not fount", mode)
	}
	return loclist, nil
}
