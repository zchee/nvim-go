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
	"time"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/cmd/guru/serial"

	"nvim-go/gb"
	"nvim-go/guru"
	"nvim-go/nvim"
)

func init() {
	plugin.HandleFunction("GoGuru", &plugin.FunctionOptions{Eval: "*"}, funcGuru)
}

type funcGuruEval struct {
	FileInfo funcGuruFileInfo
	Env      funcGuruEnv
}

type funcGuruFileInfo struct {
	Cwd  string `eval:"expand('%:p:h')"`
	File string `eval:"expand('%:p')"`
}

type funcGuruEnv struct {
	Reflection int64 `eval:"g:go#guru#reflection"`
	KeepCursor int64 `eval:"g:go#guru#keep_cursor"`
	JumpFirst  int64 `eval:"g:go#guru#jump_first"`
}

func funcGuru(v *vim.Vim, args []string, eval *funcGuruEval) {
	go Guru(v, args, eval)
}

func Guru(v *vim.Vim, args []string, eval *funcGuruEval) error {
	defer nvim.Profile(time.Now(), "Guru")

	defer gb.WithGoBuildForPath(eval.FileInfo.Cwd)()

	reflection := false
	keepCursor := false
	jumpfirst := false

	if eval.Env.Reflection == int64(1) {
		reflection = true
	}

	if eval.Env.KeepCursor == int64(1) {
		keepCursor = true
	}

	if eval.Env.JumpFirst == int64(1) {
		jumpfirst = true
	}

	var b vim.Buffer

	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	if err := p.Wait(); err != nil {
		return err
	}

	dir := strings.Split(eval.FileInfo.Cwd, "src/")
	scopeFlag := dir[len(dir)-1]

	mode := args[0]

	pos, err := nvim.ByteOffset(p)
	if err != nil {
		return nvim.Echomsg(v, "%v", err)
	}

	var outputMu sync.Mutex
	var loclist []*nvim.ErrorlistData
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		outputMu.Lock()
		defer outputMu.Unlock()
		if loclist, err = parseResult(mode, fset, qr.JSON(fset)); err != nil {
			nvim.Echoerr(v, err)
		}
	}

	query := guru.Query{
		Output:     output,
		Pos:        eval.FileInfo.File + ":#" + strconv.FormatInt(int64(pos), 10),
		Build:      &build.Default,
		Scope:      []string{scopeFlag},
		Reflection: reflection,
	}

	if err := guru.Run(mode, &query); err != nil {
		return nvim.Echomsg(v, "GoGuru:", err)
	}

	if err := nvim.SetLoclist(p, loclist); err != nil {
		return nvim.Echomsg(v, "GoGuru:", err)
	}

	if jumpfirst || mode == "definition" {
		p.Command("silent! ll 1")
		p.FeedKeys("zz", "n", false)
		return nil
	} else {
		var w vim.Window
		p.CurrentWindow(&w)
		return nvim.OpenLoclist(p, w, loclist, keepCursor)
	}
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

	case "referrers":
		var packages = serial.ReferrersPackage{}
		if err := json.Unmarshal(data, &packages); err != nil {
			return loclist, err
		}
		for _, v := range packages.Refs {
			fname, line, col := nvim.SplitPos(v.Pos)
			loclist = append(loclist, &nvim.ErrorlistData{
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
