// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"nvim-go/nvimutil"
	"nvim-go/pathutil"
	"sort"

	delveapi "github.com/derekparker/delve/service/api"
	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

func (d *Delve) printTerminal(cmd string, message []byte) error {
	d.v.SetBufferOption(d.buffer[Terminal].Buffer, "modifiable", true)
	defer d.v.SetBufferOption(d.buffer[Terminal].Buffer, "modifiable", false)

	lcount, err := d.v.BufferLineCount(d.buffer[Terminal].Buffer)
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}
	if lcount == 1 {
		lcount = 0
	}

	var msg []byte
	if cmd != "" {
		msg = []byte("(dlv) " + cmd + "\n")
		lcount--
	}
	msg = append(msg, bytes.TrimSpace(message)...)
	if len(message) != 0 {
		msg = append(msg, []byte("\n")...)
	}
	msg = append(msg, []byte("(dlv)  ")...)

	if err := d.v.SetBufferLines(d.buffer[Terminal].Buffer, lcount, -1, true, bytes.Split(msg, []byte{'\n'})); err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	afterBuf, err := d.v.BufferLines(d.buffer[Terminal].Buffer, 0, -1, true)
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	return d.v.SetWindowCursor(d.buffer[Terminal].Window, [2]int{len(afterBuf), 7})
}

func (d *Delve) printContext(cwd string, cThread *delveapi.Thread, goroutines []*delveapi.Goroutine) error {
	d.v.SetBufferOption(d.buffer[Context].Buffer, "modifiable", true)
	defer d.v.SetBufferOption(d.buffer[Context].Buffer, "modifiable", false)

	stackHeight, err := d.printStacktrace(cwd, cThread.Function, goroutines)
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	if err := d.printLocals(cwd, d.Locals, stackHeight); err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	return nil
}

// byGroutineID sorts the []*delveapi.Groutine slice by groutine ID
type byGroutineID []*delveapi.Goroutine

func (a byGroutineID) Len() int           { return len(a) }
func (a byGroutineID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byGroutineID) Less(i, j int) bool { return a[i].ID < a[j].ID }

const goroutineDepth = 20

func (d *Delve) printStacktrace(cwd string, currentFunc *delveapi.Function, goroutines []*delveapi.Goroutine) (int, error) {
	sort.Sort(byGroutineID(goroutines))

	var locals []delveapi.Variable
	var fade *nvimutil.Fade

	stacksMsg := []byte("Stacktraces\n")
	end, _ := d.v.BufferLineCount(d.buffer[Context].Buffer)

	for _, g := range goroutines {
		// Get the each threads function name.
		if g.CurrentLoc.Function.Name == currentFunc.Name {
			stacksMsg = append(stacksMsg, byte('*'))
			hlLine := len(nvimutil.ToBufferLines(stacksMsg))
			fade = nvimutil.NewFader(d.v, d.buffer[Context].Buffer, "delveFade", hlLine, hlLine, 3, -1, 80)
		} else {
			stacksMsg = append(stacksMsg, []byte(fmt.Sprintf("\t\u25B6 %s\n", g.CurrentLoc.Function.Name))...) // \u25B6: ▶
			continue
		}

		stacksMsg = append(stacksMsg, []byte(fmt.Sprintf("\t\u25BC %s\n", g.CurrentLoc.Function.Name))...) // \u25BC: ▼

		// Appends the stacktrace from each threads goroutine if valid goroutine ID.
		if g.ID != 0 {
			stacks, err := d.client.Stacktrace(g.ID, goroutineDepth, &delveapi.LoadConfig{FollowPointers: true}) // []delveapi.Stackframe
			if err != nil {
				return end, nvimutil.ErrorWrap(d.v, errors.Wrap(err, pkgDelve))
			}
			for _, s := range stacks {
				stacksMsg = append(stacksMsg, []byte(
					fmt.Sprintf("\t\t\t%s()\t%s:%d\n",
						s.Function.Name,
						pathutil.ShortFilePath(s.File, cwd),
						s.Line))...)
				locals = append(locals, s.Locals...)
			}
		}
	}

	stackData := nvimutil.ToBufferLines(stacksMsg)

	// Saves and calculates the last stacktrace message height, and check the whether the first appned to buffer.
	if end == 1 {
		end = -1
	} else {
		end = len(stackData)
	}

	if err := d.v.SetBufferLines(d.buffer[Context].Buffer, 0, end, true, stackData); err != nil {
		return end, errors.Wrap(err, pkgDelve)
	}

	if err := fade.FadeOut(); err != nil {
		return end, errors.Wrap(err, pkgDelve)
	}

	// TODO(zchee): Comparison and cacheing.
	// fmt.Sprintf("\t%s\n\t\taddr: %d\n\t\tonlyAddr: %t\n\t\ttype: %s\n\t\trealType: %s\n\t\tkind: %s\n\t\tvalue: %s\n\t\tlen: %d\n\t\tcap: %d\n\t\tunreadable: %s\n",
	// 	l.Name,
	// 	l.Addr,
	// 	l.OnlyAddr,
	// 	l.Type,
	// 	l.RealType,
	// 	l.Kind.String(),
	// 	l.Value,
	// 	l.Len,
	// 	l.Cap,
	// 	l.Unreadable))...)
	d.Locals = locals

	return end, nil
}

func (d *Delve) printLocals(cwd string, locals []delveapi.Variable, stackHeight int) error {
	localsMsg := []byte("Local Variables\n")
	for _, l := range locals {
		localsMsg = append(localsMsg, []byte(fmt.Sprintf("\t\u25B6 %s %s\n", l.Name, l.Kind.String()))...) // \u25B6: ▶

	}
	if err := d.v.SetBufferLines(d.buffer[Context].Buffer, stackHeight, -1, true, bytes.Split(localsMsg, []byte{'\n'})); err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	return nil
}

func (d *Delve) printThread(v *nvim.Nvim, cwd string, threads []*delveapi.Thread) error {
	v.SetBufferOption(d.buffer[Context].Buffer, "modifiable", true)
	defer v.SetBufferOption(d.buffer[Context].Buffer, "modifiable", false)

	for _, thread := range threads {
		printDebug("thread", thread.File)
	}

	return nil
}

func printDebug(prefix string, data interface{}) error {
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("PrintDebug: %s\n%s", prefix, data)
	}
	log.Printf("PrintDebug: %s\n%s", prefix, d)

	return nil
}
