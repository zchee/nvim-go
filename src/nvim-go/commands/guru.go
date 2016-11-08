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
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"sync"
	"time"

	"nvim-go/config"
	"nvim-go/internal/guru"
	"nvim-go/internal/guru/serial"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"

	vim "github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"github.com/ugorji/go/codec"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
)

var pkgGuru = "Guru"

type funcGuruEval struct {
	Cwd      string `msgpack:",array"`
	File     string
	Modified int
	Offset   int
}

func (c *Commands) funcGuru(args []string, eval *funcGuruEval) {
	go c.Guru(args, eval)
}

// Guru go source analysis and output result to the quickfix or locationlist.
func (c *Commands) Guru(args []string, eval *funcGuruEval) (err error) {
	defer nvimutil.Profile(time.Now(), "Guru")
	mode := args[0]
	if len(args) > 1 {
		return guruHelp(c.v, mode)
	}

	dir := filepath.Dir(eval.File)
	defer c.ctxt.SetContext(dir)()

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("guru internal panic.\nMaybe your set 'g:go#guru#reflection' to 1. Please retry with disable it option.\nOriginal panic message:\n\t%v", r.(error))
			nvimutil.ErrorWrap(c.v, err)
		}
	}()

	var (
		b vim.Buffer
		w vim.Window
	)
	if c.p == nil {
		c.p = c.v.NewPipeline()
	}
	c.p.CurrentBuffer(&b)
	c.p.CurrentWindow(&w)
	if err := c.p.Wait(); err != nil {
		return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
	}

	guruContext := &build.Default

	// https://github.com/golang/tools/blob/master/cmd/guru/main.go
	if eval.Modified != 0 {
		overlay := make(map[string][]byte)
		var buf [][]byte

		c.p.BufferLines(b, 0, -1, true, &buf)
		if err := c.p.Wait(); err != nil {
			return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
		}

		overlay[eval.File] = bytes.Join(buf, []byte{'\n'})
		guruContext = buildutil.OverlayContext(guruContext, overlay)
	}

	var loclist []*vim.QuickfixError
	query := guru.Query{
		Pos:        fmt.Sprintf("%s:#%d", eval.File, eval.Offset),
		Build:      guruContext,
		Reflection: config.GuruReflection,
	}

	if mode == "definition" {
		obj, err := definition(&query)
		if err != nil {
			return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
		}
		fname, line, col := nvimutil.SplitPos(obj.ObjPos, eval.Cwd)
		text := obj.Desc
		if fname != eval.File {
			c.p.Command(fmt.Sprintf("edit %s", fname))
		}
		c.p.SetWindowCursor(w, [2]int{line, col - 1})
		if err := c.p.Wait(); err != nil {
			return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
		}
		c.p.Command(`lclose`)
		c.p.Command(`normal! zz`)

		defer func() {
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
			nvimutil.SetLoclist(c.v, loclist)
		}()

		return c.p.Wait()
	}

	var scope string
	switch c.ctxt.Build.Tool {
	case "go":
		pkgID, err := pathutil.PackageID(dir)
		if err != nil {
			return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
		}
		scope = pkgID
	case "gb":
		scope = pathutil.GbProjectName(c.ctxt.Build.ProjectRoot)
	}
	query.Scope = append(query.Scope, filepath.Join(scope, "..."))

	var outputMu sync.Mutex
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		var err error
		outputMu.Lock()
		defer outputMu.Unlock()
		if loclist, err = parseResult(mode, fset, qr.JSON(fset), eval.Cwd); err != nil {
			nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
		}
	}
	query.Output = output

	nvimutil.EchoProgress(c.v, pkgGuru, fmt.Sprintf("analysing %s", mode))
	if err := guru.Run(mode, &query); err != nil {
		return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
	}
	if len(loclist) == 0 {
		return fmt.Errorf("%s not fount", mode)
	}

	defer nvimutil.ClearMsg(c.v)
	if err := nvimutil.SetLoclist(c.v, loclist); err != nil {
		return nvimutil.ErrorWrap(c.v, errors.Wrap(err, pkgGuru))
	}

	// jumpfirst or definition mode
	if config.GuruJumpFirst {
		c.p.Command(`silent ll 1`)
		c.p.Command(`normal! zz`)
		return c.p.Wait()
	}

	var keepCursor bool
	if int64(1) == config.GuruKeepCursor[mode] {
		keepCursor = true
	}
	return nvimutil.OpenLoclist(c.v, w, loclist, keepCursor)
}

type fallback struct {
	Obj *serial.Definition
	Err error
}

// definition reports the location of the definition of an identifier.
//
// Imported by golang.org/x/tools/cmd/guru/definition.go
// Modify use goroutine and channel.
func definition(q *guru.Query) (*serial.Definition, error) {
	defer nvimutil.Profile(time.Now(), "definition")

	c := make(chan fallback)
	go definitionFallback(q, c)

	// First try the simple resolution done by parser.
	// It only works for intra-file references but it is very fast.
	// (Extending this approach to all the files of the package,
	// resolved using ast.NewPackage, was not worth the effort.)
	qpos, err := fastQueryPos(q.Build, q.Pos)
	if err != nil {
		return nil, err
	}

	id, _ := qpos.path[0].(*ast.Ident)
	if id == nil {
		err := errors.New("no identifier here")
		return nil, err
	}

	// Did the parser resolve it to a local object?
	if obj := id.Obj; obj != nil && obj.Pos().IsValid() {
		return &serial.Definition{
			ObjPos: qpos.fset.Position(obj.Pos()).String(),
			Desc:   fmt.Sprintf("%s %s", obj.Kind, obj.Name),
		}, nil
	}

	// Qualified identifier?
	if pkg := guru.PackageForQualIdent(qpos.path, id); pkg != "" {
		srcdir := filepath.Dir(qpos.fset.File(qpos.start).Name())
		tok, pos, err := guru.FindPackageMember(q.Build, qpos.fset, srcdir, pkg, id.Name)
		if err != nil {
			return nil, err
		}
		return &serial.Definition{
			ObjPos: qpos.fset.Position(pos).String(),
			Desc:   fmt.Sprintf("%s %s.%s", tok, pkg, id.Name),
		}, nil
	}

	obj := <-c
	if obj.Err != nil {
		return nil, obj.Err
	}

	return obj.Obj, nil
}

func fallbackChan(obj *serial.Definition, err error) fallback {
	return fallback{
		Obj: obj,
		Err: err,
	}
}

// definitionFallback fall back on the type checker.
func definitionFallback(q *guru.Query, c chan fallback) {
	defer nvimutil.Profile(time.Now(), "definitionFallback")

	// Run the type checker.
	lconf := loader.Config{
		AllowErrors: true,
		Build:       q.Build,
		ParserMode:  parser.AllErrors,
		TypeChecker: types.Config{
			IgnoreFuncBodies:         true,
			FakeImportC:              true,
			DisableUnusedImportCheck: true,
			// AllErrors makes the parser always return an AST instead of
			// bailing out after 10 errors and returning an empty ast.File.
			Error: func(err error) {},
		},
	}

	if _, err := importQueryPackage(q.Pos, &lconf); err != nil {
		c <- fallbackChan(nil, err)
		return
	}

	// Load/parse/type-check the program.
	lprog, err := lconf.Load()
	if err != nil {
		c <- fallbackChan(nil, err)
		return
	}

	qpos, err := parseQueryPos(lprog, q.Pos, false)
	if err != nil {
		c <- fallbackChan(nil, err)
		return
	}

	id, ok := qpos.path[0].(*ast.Ident)
	if !ok {
		err := errors.New("no identifier here")
		c <- fallbackChan(nil, err)
		return
	}

	obj := qpos.info.ObjectOf(id)
	if obj == nil {
		err := errors.New("no object for identifier")
		c <- fallbackChan(nil, err)
		return
	}

	if !obj.Pos().IsValid() {
		err := errors.Errorf("%s is built in", obj.Name())
		c <- fallbackChan(nil, err)
		return
	}

	res := serial.Definition{
		ObjPos: qpos.fset.Position(obj.Pos()).String(),
		Desc:   qpos.ObjectString(obj),
	}

	c <- fallbackChan(&res, nil)
	return
}

var null = []byte{110, 117, 108, 108}

// TODO(zchee): Should not use json.
func parseResult(mode string, fset *token.FileSet, data []byte, cwd string) ([]*vim.QuickfixError, error) {
	if nvimutil.IsDebug() {
		log.Printf("data: %+s", string(data))
	}
	var (
		loclist []*vim.QuickfixError
		fname   string
		line    int
		col     int
		text    string
	)
	var (
		jh  codec.JsonHandle
		dec = codec.NewDecoderBytes(data, &jh)
	)

	switch mode {

	case "callees":
		var value = serial.Callees{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		for _, v := range value.Callees {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = value.Desc + ": " + v.Name
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callers":
		if bytes.Equal(data, null) {
			err := errors.New("not reachable in this program")
			return nil, err
		}
		var value = []serial.Caller{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		for _, v := range value {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = v.Desc + ": " + v.Caller
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callstack":
		var value = serial.CallStack{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		for _, v := range value.Callers {
			fname, line, col = nvimutil.SplitPos(v.Pos, cwd)
			text = v.Desc + " " + value.Target
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "describe":
		var value = serial.Describe{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		fname, line, col = nvimutil.SplitPos(value.Value.ObjPos, cwd)
		text = value.Desc + " " + value.Value.Type
		loclist = append(loclist, &vim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "freevars":
		var value = serial.FreeVar{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		fname, line, col = nvimutil.SplitPos(value.Pos, cwd)
		text = value.Kind + " " + value.Type + " " + value.Ref
		loclist = append(loclist, &vim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     text,
		})

	case "implements":
		var typ = serial.Implements{}
		err := dec.Decode(&typ)
		if err != nil {
			return loclist, err
		}
		for _, value := range typ.AssignableFrom {
			fname, line, col := nvimutil.SplitPos(value.Pos, cwd)
			text = value.Kind + " " + value.Name
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}
		for _, value := range typ.AssignableFromPtr {
			fname, line, col := nvimutil.SplitPos(value.Pos, cwd)
			text = value.Kind + " " + value.Name
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "peers":
		var value = serial.Peers{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		fname, line, col := nvimutil.SplitPos(value.Pos, cwd)
		loclist = append(loclist, &vim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Base: selected channel op (<-)",
		})
		for _, v := range value.Allocs {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Allocs: make(chan) ops",
			})
		}
		for _, v := range value.Sends {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Sends: ch<-x ops",
			})
		}
		for _, v := range value.Receives {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Receives: <-ch ops",
			})
		}
		for _, v := range value.Closes {
			fname, line, col := nvimutil.SplitPos(v, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Closes: close(ch) ops",
			})
		}

	case "pointsto":
		var value = []serial.PointsTo{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		for _, v := range value {
			fname, line, col := nvimutil.SplitPos(v.NamePos, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "type: " + v.Type,
			})
			if len(v.Labels) > -1 {
				for _, vl := range v.Labels {
					fname, line, col := nvimutil.SplitPos(vl.Pos, cwd)
					loclist = append(loclist, &vim.QuickfixError{
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
		if err := dec.Decode(&packages); err != nil {
			return loclist, err
		}
		for _, v := range packages.Refs {
			fname, line, col := nvimutil.SplitPos(v.Pos, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     v.Text,
			})
		}

	case "whicherrs":
		var value = serial.WhichErrs{}
		err := dec.Decode(&value)
		if err != nil {
			return loclist, err
		}
		fname, line, col := nvimutil.SplitPos(value.ErrPos, cwd)
		loclist = append(loclist, &vim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Errror Position",
		})
		for _, vg := range value.Globals {
			fname, line, col := nvimutil.SplitPos(vg, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Globals",
			})
		}
		for _, vc := range value.Constants {
			fname, line, col := nvimutil.SplitPos(vc, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Constants",
			})
		}
		for _, vt := range value.Types {
			fname, line, col := nvimutil.SplitPos(vt.Position, cwd)
			loclist = append(loclist, &vim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     "Types: " + vt.Type,
			})
		}

	}
	return loclist, nil
}

func guruHelp(v *vim.Nvim, mode string) error {
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
