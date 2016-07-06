// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/pathutil"

	delveapi "github.com/derekparker/delve/service/api"
	delverpc2 "github.com/derekparker/delve/service/rpc2"
	delveterm "github.com/derekparker/delve/terminal"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
)

const (
	addr string = "localhost:41222" // d:4 l:12 v:22

	pkgDelve string = "Delve"
)

func init() {
	d := NewDelve()

	// Launch
	plugin.HandleCommand("DlvDebug", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdDebug) // Compile and begin debugging program.

	// Breakpoint
	plugin.HandleCommand("DlvBreakpoint", &plugin.CommandOptions{NArgs: "*", Eval: "[expand('%:p')]", Complete: "customlist,DlvListFunctions"}, d.cmdCreateBreakpoint) // Sets a breakpoint.

	// Stepping execution control
	plugin.HandleCommand("DlvContinue", &plugin.CommandOptions{Eval: "[expand('%:p:h')]"}, d.cmdContinue) // Run until breakpoint or program termination.
	plugin.HandleCommand("DlvNext", &plugin.CommandOptions{Eval: "[expand('%:p:h')]"}, d.cmdNext)         // Step over to next source line.
	plugin.HandleCommand("DlvRestart", &plugin.CommandOptions{}, d.cmdRestart)                            // Restart process.

	// Interactive mode
	// XXX(zchee): Support contextual command completion
	plugin.HandleCommand("DlvStdin", &plugin.CommandOptions{}, d.cmdStdin)
	plugin.HandleFunction("DlvListFunctions", &plugin.FunctionOptions{}, d.ListFunctions) // list of functions for command completion.

	// Detach
	plugin.HandleCommand("DlvDetach", &plugin.CommandOptions{}, d.cmdDetach) // Exit the debugger.

	// RPC Exports
	plugin.Handle("DlvStdin", d.stdin)

	// State (WIP: for debug)
	plugin.HandleCommand("DlvState", &plugin.CommandOptions{}, d.cmdState)

	// autocmd VimLeavePre
	// FIXME(zchee): Why "[delve]*" pattern dose not handle autocmd?
	plugin.HandleAutocmd("VimLeavePre", &plugin.AutocmdOptions{Group: "nvim-go", Pattern: "*.go,terminal,context,thread"}, d.cmdDetach)
}

// Delve represents a delve client.
type Delve struct {
	v *vim.Vim

	server     *exec.Cmd
	client     *delverpc2.RPCClient
	term       *delveterm.Term
	debugger   *delveterm.Commands
	processPid int
	serverOut  bytes.Buffer
	serverErr  bytes.Buffer

	channelID int

	Locals []delveapi.Variable

	BufferContext
	SignContext

	ctxt *context.Context
}

// BufferContext represents a each debug information buffers.
type BufferContext struct {
	cb     vim.Buffer
	cw     vim.Window
	buffer map[string]*nvim.Buf
}

// SignContext represents a breakpoint and program counter sign.
type SignContext struct {
	bpSign map[int]*nvim.Sign // map[breakPoint.id]*nvim.Sign
	pcSign *nvim.Sign
}

// NewDelve represents a delve client interface.
func NewDelve() *Delve {
	return &Delve{}
}

// setupDelveClient setup the delve client. Separate the NewDelveClient() function.
// caused by neovim-go can't call the rpc2.NewClient?
func (d *Delve) setupDelve(v *vim.Vim) error {
	d.v = v

	d.client = delverpc2.NewClient(addr)           // *rpc2.RPCClient
	d.term = delveterm.New(d.client, nil)          // *terminal.Term
	d.debugger = delveterm.DebugCommands(d.client) // *terminal.Commands
	d.processPid = d.client.ProcessPid()           // int

	return nil
}

// debugEval represent a debug commands Eval args.
type debugEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

func (d *Delve) cmdDebug(v *vim.Vim, eval *debugEval) {
	d.debug(v, eval)
}

// TODO(zchee): If failed debug(build), even create each buffers.
func (d *Delve) debug(v *vim.Vim, eval *debugEval) error {
	d.ctxt = new(context.Context)
	defer d.ctxt.Build.SetContext(eval.Cwd)()

	rootDir := pathutil.FindVcsRoot(eval.Dir)
	srcPath := filepath.Join(os.Getenv("GOPATH"), "src") + string(filepath.Separator)
	path := filepath.Clean(strings.TrimPrefix(rootDir, srcPath))

	if err := d.startServer("debug", path); err != nil {
		nvim.ErrorWrap(v, err)
	}
	defer d.waitServer(v)

	return d.createDebugBuffer(v)
}

func (d *Delve) parseArgs(v *vim.Vim, args []string, eval *createBreakpointEval) (*delveapi.Breakpoint, error) {
	var bpInfo *delveapi.Breakpoint

	// TODO(zchee): Now support function only.
	// Ref: https://github.com/derekparker/delve/blob/master/Documentation/cli/locspec.md
	switch len(args) {
	case 0:
		cursor, err := v.WindowCursor(d.cw)
		if err != nil {
			return nil, err
		}

		bpInfo = &delveapi.Breakpoint{
			File: eval.File,
			Line: cursor[0],
		}
	case 1:
		// FIXME(zchee): more elegant way
		splitargs := strings.Split(args[0], ".")
		splitargs[1] = fmt.Sprintf("%s%s", strings.ToUpper(splitargs[1][:1]), splitargs[1][1:])
		name := strings.Join(splitargs, "")

		bpInfo = &delveapi.Breakpoint{
			Name:         name,
			FunctionName: args[0],
		}
	default:
		return nil, errors.Annotate(errors.New("Too many arguments"), pkgDelve)
	}

	return bpInfo, nil
}

// breakpointEval represent a breakpoint commands Eval args.
type createBreakpointEval struct {
	File string `msgpack:",array"`
}

func (d *Delve) cmdCreateBreakpoint(v *vim.Vim, args []string, eval *createBreakpointEval) {
	go d.createBreakpoint(v, args, eval)
}

func (d *Delve) createBreakpoint(v *vim.Vim, args []string, eval *createBreakpointEval) error {
	bpInfo, err := d.parseArgs(v, args, eval)
	if err != nil {
		nvim.ErrorWrap(v, err)
	}

	if d.bpSign == nil {
		d.bpSign = make(map[int]*nvim.Sign)
	}

	bp, err := d.client.CreateBreakpoint(bpInfo) // *delveapi.Breakpoint
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	d.bpSign[bp.ID], err = nvim.NewSign(v, "delve_bp", nvim.BreakpointSymbol, "delveBreakpointSign", "") // *nvim.Sign
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}
	d.bpSign[bp.ID].Place(v, bp.ID, bp.Line, bp.File, false)

	if err := d.printTerminal(v, "break "+bp.FunctionName, []byte{}); err != nil {
		return nvim.ErrorWrap(v, err)
	}

	return nil
}

// breakpointEval represent a breakpoint commands Eval args.
type continueEval struct {
	Dir string `msgpack:",array"`
}

func (d *Delve) cmdContinue(v *vim.Vim, eval *continueEval) {
	go d.cont(v, eval)
}

// cont sends the 'continue' signals to the delve headless server over the client use json-rpc2 protocol.
func (d *Delve) cont(v *vim.Vim, eval *continueEval) error {
	stateCh := d.client.Continue()
	state := <-stateCh
	if state == nil || state.Exited {
		return nvim.ErrorWrap(v, errors.Annotate(state.Err, pkgDelve))
	}

	cThread := state.CurrentThread

	go func() {
		goroutines, err := d.client.ListGoroutines()
		if err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
		d.printContext(v, eval.Dir, cThread, goroutines)
	}()

	go d.pcSign.Place(v, cThread.ID, cThread.Line, cThread.File, true)

	go func() {
		if err := v.SetWindowCursor(d.cw, [2]int{cThread.Line, 0}); err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
		if err := v.Command("silent normal zz"); err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
	}()

	var msg []byte
	if hitCount, ok := cThread.Breakpoint.HitCount[strconv.Itoa(cThread.GoroutineID)]; ok {
		msg = []byte(
			fmt.Sprintf("> %s() %s:%d (hits goroutine(%d):%d total:%d) (PC: %#v)",
				cThread.Function.Name,
				shortFilePath(cThread.File, eval.Dir),
				cThread.Line,
				cThread.GoroutineID,
				hitCount,
				cThread.Breakpoint.TotalHitCount,
				cThread.PC))
	} else {
		msg = []byte(
			fmt.Sprintf("> %s() %s:%d (hits total:%d) (PC: %#v)",
				cThread.Function.Name,
				shortFilePath(cThread.File, eval.Dir),
				cThread.Line,
				cThread.Breakpoint.TotalHitCount,
				cThread.PC))
	}
	return d.printTerminal(v, "continue", msg)
}

// breakpointEval represent a breakpoint commands Eval args.
type nextEval struct {
	Dir string `msgpack:",array"`
}

func (d *Delve) cmdNext(v *vim.Vim, eval *nextEval) {
	go d.next(v, eval)
}

func (d *Delve) next(v *vim.Vim, eval *nextEval) error {
	state, err := d.client.Next()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	cThread := state.CurrentThread

	go func() {
		goroutines, err := d.client.ListGoroutines()
		if err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
		d.printContext(v, eval.Dir, cThread, goroutines)
	}()

	go d.pcSign.Place(v, cThread.ID, cThread.Line, cThread.File, true)

	go func() {
		if err := v.SetWindowCursor(d.cw, [2]int{cThread.Line, 0}); err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
		if err := v.Command("silent normal zz"); err != nil {
			nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
			return
		}
	}()

	msg := []byte(
		fmt.Sprintf("> %s() %s:%d goroutine(%d) (PC: %d)",
			cThread.Function.Name,
			shortFilePath(cThread.File, eval.Dir),
			cThread.Line,
			cThread.GoroutineID,
			cThread.PC))
	return d.printTerminal(v, "next", msg)
}

func (d *Delve) cmdRestart(v *vim.Vim) {
	go d.restart(v)
}

func (d *Delve) restart(v *vim.Vim) error {
	err := d.client.Restart()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve+"restart"))
	}

	d.processPid = d.client.ProcessPid()
	return d.printTerminal(v, "restart", []byte(fmt.Sprintf("Process restarted with PID %d", d.processPid)))
}

func (d *Delve) cmdStdin(v *vim.Vim) {
	go d.stdin(v)
}

// ListFunctions return the debug target functions with filtering "main".
func (d *Delve) ListFunctions(v *vim.Vim) ([]string, error) {
	funcs, err := d.client.ListFunctions("main")
	if err != nil {
		return []string{}, err
	}

	return funcs, nil
}

// stdin sends the users input command to the internal delve terminal.
// vim input() function args:
//  input({prompt} [, {text} [, {completion}]])
// More information of input() funciton and word completion are
//  :help input()
//  :help command-completion-custom
func (d *Delve) stdin(v *vim.Vim) error {
	var stdin interface{}
	err := v.Call("input", &stdin, "(dlv) ", "")
	if err != nil {
		return nil
	}

	// Create the connected pair of *os.Files and replace os.Stdout.
	// delve terminal package return to stdout only.
	r, w, _ := os.Pipe() // *os.File
	saveStdout := os.Stdout
	os.Stdout = w

	cmd := strings.SplitN(stdin.(string), " ", 2)
	var args string
	if len(cmd) == 2 {
		args = cmd[1]
	}

	err = d.debugger.Call(cmd[0], args, d.term)
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	// Close the w file and restore os.Stdout to original.
	w.Close()
	os.Stdout = saveStdout

	// Read all the lines of r file.
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	return d.printTerminal(v, stdin.(string), out)
}

func shortFilePath(p, cwd string) string {
	return strings.Replace(p, cwd, ".", 1)
}

func (d *Delve) cmdState(v *vim.Vim) {
	go d.state(v)
}

func (d *Delve) state(v *vim.Vim) error {
	state, err := d.client.GetState()
	if err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	printDebug("state: %+v\n", state)
	return nil
}
