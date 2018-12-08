// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// guru: a tool for answering questions about Go source code.
//
//    http://golang.org/s/using-guru

package command

import (
	"bytes"
	"context"
	"fmt"
	"go/build"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"golang.org/x/tools/cmd/guru/serial"
	"golang.org/x/tools/go/buildutil"

	"github.com/zchee/nvim-go/pkg/config"
	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/internal/guru"
	"github.com/zchee/nvim-go/pkg/logger"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

type funcGuruEval struct {
	Cwd      string `msgpack:",array"`
	File     string
	Modified int
	Offset   int
}

func (c *Command) funcGuru(args []string, eval *funcGuruEval) {
	errch := make(chan interface{}, 1)
	go func() {
		errch <- c.Guru(c.ctx, args, eval)
	}()

	select {
	case <-c.ctx.Done():
		return
	case err := <-errch:
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Guru", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		case nil:
			// nothing to do
		}
	}
}

// Guru go source analysis and output result to the quickfix or locationlist.
func (c *Command) Guru(ctx context.Context, args []string, eval *funcGuruEval) interface{} {
	defer nvimutil.Profile(ctx, time.Now(), "Guru")
	span := trace.FromContext(ctx)
	span.SetName("Guru")
	defer span.End()

	log := logger.FromContext(c.ctx).Named("Guru").With(zap.Any("funcGuruEval", eval))

	mode := args[0]
	if len(args) > 1 {
		return guruHelp(c.Nvim, mode)
	}

	defer func() (err error) {
		switch r := recover(); r.(type) {
		case error:
			const errGuruPanic = "guru internal panic.\nMaybe your set 'g:go#guru#reflection' to 1. Please retry with disable it option.\nOriginal panic message:\n\t%v"
			err = errors.Errorf(errGuruPanic, r)
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		case runtime.Error:
			err = errors.Errorf("runtime error: %v", r)
			panic(err)
		}
		return nil
	}()

	b := nvim.Buffer(c.buildContext.BufNr)
	w := nvim.Window(c.buildContext.WinID)
	batch := c.Nvim.NewBatch()

	guruContext := &build.Default

	// https://github.com/golang/tools/blob/master/cmd/guru/main.go
	if eval.Modified != 0 {
		overlay := make(map[string][]byte)
		var buf [][]byte

		batch.BufferLines(b, 0, -1, true, &buf)
		if err := batch.Execute(); err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
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
	log.Info("", zap.String("query.Pos", query.Pos), zap.Bool("query.Reflection", query.Reflection))

	if mode == "definition" {
		obj, err := Definition(&query)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		}
		fname, line, col := nvimutil.SplitPos(obj.ObjPos, eval.Cwd)

		batch.Command("normal! m'")
		// TODO(zchee): should change nvimutil.SplitPos behavior
		filename := strings.Split(obj.ObjPos, ":")
		if filename[0] != eval.File {
			batch.Command(fmt.Sprintf("keepjumps edit %s", fs.Rel(eval.Cwd, fname)))
		}
		batch.SetWindowCursor(w, [2]int{line, col - 1})
		if err := batch.Execute(); err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.WithStack(err)
		}

		return c.Nvim.Command(`lclose | normal! zz`)
	}

	var scopes []string
	switch c.buildContext.Build.Tool {
	case "go":
		root := fs.FindVCSRoot(eval.File)
		root, _ = filepath.Abs(root)
		scopes = []string{fs.ToWildcard(fs.TrimGoPath(root))}
		if vendorDir := filepath.Join(root, "vendor"); fs.IsDirExist(vendorDir) {
			scopes = append(scopes, "-"+fs.TrimGoPath(vendorDir))
		}
		os.Unsetenv("GO111MODULE")
	case "gb":
		root := c.buildContext.Build.ProjectRoot
		var err error
		scopes, err = fs.GbPackages(root)
		if err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			return errors.Wrap(err, "could not get gb packages")
		}
		for i, pkg := range scopes {
			scopes[i] = fs.ToWildcard(pkg)
		}
		if vendorDir := filepath.Join(root, "vendor"); fs.IsDirExist(vendorDir) {
			scopes = append(scopes, "-"+fs.ToWildcard(vendorDir))
		}
	}
	query.Scope = append(query.Scope, scopes...)
	log.Info("",
		zap.Strings("scopes", scopes),
		zap.String("query.Pos", query.Pos),
		zap.Bool("query.Reflection", query.Reflection),
		zap.Strings("query.Scope", query.Scope))

	var outputMu sync.Mutex
	var err error
	output := func(fset *token.FileSet, qr guru.QueryResult) {
		var err error
		outputMu.Lock()
		defer outputMu.Unlock()

		res := qr.Result(fset)
		if loclist, err = c.parseResult(ctx, mode, res, eval.Cwd); err != nil {
			span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
			err = errors.WithStack(err)
		}
	}
	if err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	query.Output = output

	nvimutil.EchoProgress(c.Nvim, "Guru", fmt.Sprintf("analysing %s", mode))
	if err := guru.Run(mode, &query); err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}
	if len(loclist) == 0 {
		err := errors.Errorf("%s not found", mode)
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return err
	}

	defer nvimutil.ClearMsg(c.Nvim)
	if err := nvimutil.SetLoclist(c.Nvim, loclist); err != nil {
		span.SetStatus(trace.Status{Code: trace.StatusCodeInternal, Message: err.Error()})
		return errors.WithStack(err)
	}

	// jumpfirst or definition mode
	if config.GuruJumpFirst {
		batch.Command(`silent ll 1`)
		batch.Command(`normal! zz`)
		return batch.Execute()
	}

	var keepCursor bool
	if config.GuruKeepCursor[mode] {
		keepCursor = true
	}
	return nvimutil.OpenLoclist(c.Nvim, w, loclist, keepCursor)
}

var errTypeAssertion = errors.New("type assertion error")

func (c *Command) parseResult(ctx context.Context, mode string, res interface{}, cwd string) ([]*nvim.QuickfixError, error) {
	log := logger.FromContext(ctx).With(zap.String("mode", mode), zap.String("cwd", cwd))
	log.Info("", zap.Any("res", res))

	var loclist []*nvim.QuickfixError

	switch mode {
	case "callees":
		v, ok := res.(*serial.Callees)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, cle := range v.Callees {
			fname, line, col := nvimutil.SplitPos(cle.Pos, cwd)
			text := v.Desc + ": " + cle.Name
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callers":
		v, ok := res.([]serial.Caller)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, clr := range v {
			fname, line, col := nvimutil.SplitPos(clr.Pos, cwd)
			text := clr.Desc + ": " + clr.Caller
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "callstack":
		v, ok := res.(*serial.CallStack)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, clr := range v.Callers {
			fname, line, col := nvimutil.SplitPos(clr.Pos, cwd)
			text := clr.Desc + " " + v.Target
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "describe":
		v, ok := res.(*serial.Describe)
		if !ok {
			return loclist, errTypeAssertion
		}
		switch {
		case v.Package != nil:
			log.Info("value.Package")
		case v.Type != nil:
			log.Info("value.Type")
			for _, method := range v.Type.Methods {
				fname, line, col := nvimutil.SplitPos(method.Pos, cwd)
				text := method.Name
				log.Info("", zap.String("fname", fname), zap.Int("line", line), zap.Int("col", col), zap.String("text", text))
				loclist = append(loclist, &nvim.QuickfixError{
					FileName: fname,
					LNum:     line,
					Col:      col,
					Text:     text,
				})
			}
		case v.Value != nil:
			log.Info("value.Value")
			fname, line, col := nvimutil.SplitPos(v.Value.ObjPos, cwd)
			text := v.Desc + " " + v.Value.Type
			log.Info("", zap.String("fname", fname), zap.Int("line", line), zap.Int("col", col), zap.String("text", text))
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "freevars":
		v, ok := res.([]serial.FreeVar)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, fv := range v {
			fname, line, col := nvimutil.SplitPos(fv.Pos, cwd)
			text := fv.Kind + " " + fv.Type + " " + fv.Ref
			loclist = append(loclist, &nvim.QuickfixError{
				FileName: fname,
				LNum:     line,
				Col:      col,
				Text:     text,
			})
		}

	case "implements":
		v, ok := res.(*serial.Implements)
		if !ok {
			return loclist, errTypeAssertion
		}
		for _, tt := range [][]serial.ImplementsType{v.AssignableTo, v.AssignableFromPtr, v.AssignableFrom} {
			for _, t := range tt {
				fname, line, col := nvimutil.SplitPos(t.Pos, cwd)
				text := t.Kind + " " + t.Name
				loclist = append(loclist, &nvim.QuickfixError{
					FileName: fname,
					LNum:     line,
					Col:      col,
					Text:     text,
				})
			}
		}

	case "peers":
		vp, ok := res.(*serial.Peers)
		if !ok {
			return loclist, errTypeAssertion
		}
		fname, line, col := nvimutil.SplitPos(vp.Pos, cwd)
		loclist = append(loclist, &nvim.QuickfixError{
			FileName: fname,
			LNum:     line,
			Col:      col,
			Text:     "Base: selected channel op (<-)",
		})
		peertext := []string{
			"Allocs: make(chan) ops",
			"Sends: ch<-x ops",
			"Receives: <-ch ops",
			"Closes: close(ch) ops",
		}
		for i, vv := range [][]string{vp.Allocs, vp.Sends, vp.Receives, vp.Closes} {
			for _, v := range vv {
				fname, line, col := nvimutil.SplitPos(v, cwd)
				loclist = append(loclist, &nvim.QuickfixError{
					FileName: fname,
					LNum:     line,
					Col:      col,
					Text:     peertext[i],
				})
			}
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
