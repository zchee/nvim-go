// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// guru: a tool for answering questions about Go source code.
//
//    http://golang.org/s/oracle-design
//    http://golang.org/s/oracle-user-manual

package commands

import (
	"encoding/json"
	"fmt"
	"go/build"
	"go/token"
	"strconv"
	"strings"
	"sync"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/cmd/guru/serial"

	"nvim-go/gb"
	"nvim-go/guru"
	"nvim-go/nvim"
)

func init() {
	evalArg := "[expand('%:p:h'), expand('%:p')]"

	plugin.HandleCommand("GoGuru",
		&plugin.CommandOptions{
			NArgs: "+", Complete: "customlist,GuruCompletelist", Eval: evalArg}, cmdGuru)
	plugin.HandleFunction("GuruCompletelist", &plugin.FunctionOptions{}, onComplete)

	// Show possible callees of the function call at the current point.
	plugin.HandleCommand("GoGuruCallees", &plugin.CommandOptions{Eval: evalArg}, cmdGuruCallees)
	// Show the set of callers of the function containing the current point.
	plugin.HandleCommand("GoGuruCallers", &plugin.CommandOptions{Eval: evalArg}, cmdGuruCallers)
	// Show the callgraph of the current program.
	plugin.HandleCommand("GoGuruCallstack", &plugin.CommandOptions{Eval: evalArg}, cmdGuruCallstack)
	plugin.HandleCommand("GoGuruDefinition", &plugin.CommandOptions{Eval: evalArg}, cmdGuruDefinition)
	// Describe the expression at the current point.
	plugin.HandleCommand("GoGuruDescribe", &plugin.CommandOptions{Eval: evalArg}, cmdGuruDescribe)
	plugin.HandleCommand("GoGuruFreevars", &plugin.CommandOptions{Eval: evalArg}, cmdGuruFreevars)
	/// Describe the 'implements' relation for types in the
	// package containing the current point.
	plugin.HandleCommand("GoGuruImplements", &plugin.CommandOptions{Eval: evalArg}, cmdGuruImplements)
	// Enumerate the set of possible corresponding sends/receives for
	// this channel receive/send operation.
	plugin.HandleCommand("GoGuruChannelPeers", &plugin.CommandOptions{Eval: evalArg}, cmdGuruChannelPeers)
	plugin.HandleCommand("GoGuruPointsto", &plugin.CommandOptions{Eval: evalArg}, cmdGuruPointsto)
	plugin.HandleCommand("GoGuruWhicherrs", &plugin.CommandOptions{Eval: evalArg}, cmdGuruWhicherrs)
}

var (
	guruReflection  = "go#guru#reflection"
	vGuruReflection interface{}
	guruKeepCursor  = "go#guru#keep_cursor"
	vGuruKeepCursor interface{}
	guruDebug       = "go#debug"
	vGuruDebug      interface{}
)

type onGuruEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func cmdGuru(v *vim.Vim, args []string, eval *onGuruEval) {
	go Guru(v, args, eval)
}

func cmdGuruCallees(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"callees"}, eval)
}

func cmdGuruCallers(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"callers"}, eval)
}

func cmdGuruCallstack(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"callstack"}, eval)
}

func cmdGuruDefinition(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"definition"}, eval)
}

func cmdGuruDescribe(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"describe"}, eval)
}

func cmdGuruFreevars(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"freevars"}, eval)
}

func cmdGuruImplements(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"implements"}, eval)
}

func cmdGuruChannelPeers(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"peers"}, eval)
}

func cmdGuruPointsto(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"pointsto"}, eval)
}

func cmdGuruWhicherrs(v *vim.Vim, eval *onGuruEval) {
	go Guru(v, []string{"whicherrs"}, eval)
}

func Guru(v *vim.Vim, args []string, eval *onGuruEval) error {
	defer gb.WithGoBuildForPath(eval.Cwd)()

	useReflection := false
	v.Var(guruReflection, &vGuruReflection)
	if vGuruReflection.(int64) == int64(1) {
		useReflection = true
	}
	useKeepCursor := false
	v.Var(guruKeepCursor, &vGuruKeepCursor)
	if vGuruKeepCursor.(int64) == int64(1) {
		useKeepCursor = true
	}

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

	dir := strings.Split(eval.Cwd, "src/")
	scopeFlag := dir[len(dir)-1]

	mode := args[0]

	pos, err := nvim.ByteOffset(p)
	if err != nil {
		return nvim.Echomsg(v, "%v", err)
	}

	var outputMu sync.Mutex
	var d []*nvim.ErrorlistData
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		outputMu.Lock()
		defer outputMu.Unlock()
		if d, err = parseResult(mode, fset, qr.JSON(fset)); err != nil {
			nvim.Echoerr(v, err)
		}
	}

	query := guru.Query{
		Output:     output,
		Pos:        eval.File + ":#" + strconv.FormatInt(int64(pos), 10),
		Build:      &build.Default,
		Scope:      []string{scopeFlag},
		Reflection: useReflection,
	}

	nvim.Echohl(v, "GoGuru: ", "Identifier", "analysing %s ...", mode)

	if err := guru.Run(mode, &query); err != nil {
		return nvim.Echomsg(v, err)
	}
	v.Command("normal :<ESC>")

	if err := nvim.SetLoclist(p, d); err != nil {
		return nvim.Echomsg(v, "GoGuru: %v", err)
	}
	return nvim.OpenLoclist(p, w, d, useKeepCursor)
}

func parseResult(mode string, fset *token.FileSet, data []byte) ([]*nvim.ErrorlistData, error) {
	var (
		loclist []*nvim.ErrorlistData
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
			fname, line, col = nvim.SplitPos(v.Pos)
			text = value.Desc + ": " + v.Name
			loclist = append(loclist, &nvim.ErrorlistData{
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
			fname, line, col = nvim.SplitPos(v.Pos)
			text = v.Desc + ": " + v.Caller
			loclist = append(loclist, &nvim.ErrorlistData{
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
			fname, line, col = nvim.SplitPos(v.Pos)
			text = v.Desc + " " + value.Target
			loclist = append(loclist, &nvim.ErrorlistData{
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
		fname, line, col = nvim.SplitPos(value.ObjPos)
		text = value.Desc
		loclist = append(loclist, &nvim.ErrorlistData{
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
		fname, line, col = nvim.SplitPos(value.Value.ObjPos)
		text = value.Desc + " " + value.Value.Type
		loclist = append(loclist, &nvim.ErrorlistData{
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
		fname, line, col = nvim.SplitPos(value.Pos)
		text = value.Kind + " " + value.Type + " " + value.Ref
		loclist = append(loclist, &nvim.ErrorlistData{
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
			fname, line, col := nvim.SplitPos(v.Pos)
			text = v.Kind + " " + v.Name
			loclist = append(loclist, &nvim.ErrorlistData{
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
		fname, line, col := nvim.SplitPos(value.Pos)
		loclist = append(loclist, &nvim.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Base: selected channel op (<-)",
		})
		for _, v := range value.Allocs {
			fname, line, col := nvim.SplitPos(v)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Allocs: make(chan) ops",
			})
		}
		for _, v := range value.Sends {
			fname, line, col := nvim.SplitPos(v)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Sends: ch<-x ops",
			})
		}
		for _, v := range value.Receives {
			fname, line, col := nvim.SplitPos(v)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Receives: <-ch ops",
			})
		}
		for _, v := range value.Closes {
			fname, line, col := nvim.SplitPos(v)
			loclist = append(loclist, &nvim.ErrorlistData{
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
			fname, line, col := nvim.SplitPos(v.NamePos)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "type: " + v.Type,
			})
			if len(v.Labels) > -1 {
				for _, vl := range v.Labels {
					fname, line, col := nvim.SplitPos(vl.Pos)
					loclist = append(loclist, &nvim.ErrorlistData{
						FileName: fname,
						LNum:     line,
						Col:      col,
						Text:     vl.Desc,
					})
				}
			}
		}

	// case "referrers":
	// 	var value = serial.ReferrersInitial{}
	// 	err := json.Unmarshal(data, &value)
	// 	if err != nil {
	// 		return loclist, err
	// 	}
	// 	fname, line, col := nvim.SplitPos(value.ObjPos)
	// 	loclist = append(loclist, &nvim.ErrorlistData{
	// 		FileName: fname,
	// 		LNum:     line,
	// 		Col:      col,
	// 		Text:     value.Desc,
	// 	})
	// 	var vPackage = serial.ReferrersPackage{}
	// 	if err := json.Unmarshal(data, &vPackage); err != nil {
	// 		return loclist, err
	// 	}
	// 	loclist = append(loclist, &nvim.ErrorlistData{
	// 		Text: vPackage.Package,
	// 	})
	// 	for _, vp := range vPackage.Refs {
	// 		fname, line, col := nvim.SplitPos(vp.Pos)
	// 		loclist = append(loclist, &nvim.ErrorlistData{
	// 			FileName: fname,
	// 			LNum:     line,
	// 			Col:      col,
	// 			Text:     vp.Text,
	// 		})
	// 	}

	case "whicherrs":
		var value = serial.WhichErrs{}
		err := json.Unmarshal(data, &value)
		if err != nil {
			return loclist, err
		}
		fname, line, col := nvim.SplitPos(value.ErrPos)
		loclist = append(loclist, &nvim.ErrorlistData{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Errror Position",
		})
		for _, vg := range value.Globals {
			fname, line, col := nvim.SplitPos(vg)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Globals",
			})
		}
		for _, vc := range value.Constants {
			fname, line, col := nvim.SplitPos(vc)
			loclist = append(loclist, &nvim.ErrorlistData{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Constants",
			})
		}
		for _, vt := range value.Types {
			fname, line, col := nvim.SplitPos(vt.Position)
			loclist = append(loclist, &nvim.ErrorlistData{
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

func onComplete(v *vim.Vim) ([]string, error) {
	return []string{
		"callees",
		"callers",
		"callstack",
		"definition",
		"describe",
		"freevars",
		"implements",
		"peers",
		"pointsto",
		"referrers",
		"what",
		"whicherrs",
	}, nil
}
