package delve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nvim-go/context"
	"nvim-go/nvim"
	"nvim-go/nvim/buffer"

	delveapi "github.com/derekparker/delve/service/api"
	delverpc2 "github.com/derekparker/delve/service/rpc2"
	delveterm "github.com/derekparker/delve/terminal"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/juju/errors"
)

const (
	addr string = "localhost:41222" // d:4 l:12 v:22

	pkgDelve string = "Delve."
)

func init() {
	d := NewDelveClient()
	// Launch
	plugin.HandleCommand("DlvDebug", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdDebug)

	// Breakpoint
	plugin.HandleCommand("DlvBreakpoint", &plugin.CommandOptions{NArgs: "*", Eval: "[expand('%:p')]", Complete: "customlist,DelveFunctionList"}, d.cmdCreateBreakpoint)

	// Stepping execution control
	plugin.HandleCommand("DlvContinue", &plugin.CommandOptions{Eval: "[expand('%:p:h')]"}, d.cmdContinue)
	plugin.HandleCommand("DlvNext", &plugin.CommandOptions{Eval: "[expand('%:p:h')]"}, d.cmdNext)
	plugin.HandleCommand("DlvRestart", &plugin.CommandOptions{}, d.cmdRestart)

	// Interactive mode
	// XXX(zchee): Support contextual command completion
	plugin.HandleCommand("DlvStdin", &plugin.CommandOptions{}, d.stdin)
	plugin.HandleFunction("DelveListFunctions", &plugin.FunctionOptions{}, d.ListFunctions)

	// Detach
	plugin.HandleCommand("DlvDetach", &plugin.CommandOptions{}, d.cmdDetach)

	// RPC Exports
	plugin.Handle("DlvStdin", d.stdin)

	// State (WIP: for debug)
	plugin.HandleCommand("DlvState", &plugin.CommandOptions{}, d.cmdState)

	// autocmd VimLeavePre
	// FIXME(zchee): Why "[delve]*" pattern dose not handle autocmd?
	plugin.HandleAutocmd("VimLeavePre", &plugin.AutocmdOptions{Group: "nvim-go", Pattern: "*.go,terminal,stacktrace,locals,threads"}, d.cmdDetach)
}

type delveClient struct {
	server   *exec.Cmd
	client   *delverpc2.RPCClient
	term     *delveterm.Term
	debugger *delveterm.Commands

	processPid           int
	serverOut, serverErr bytes.Buffer

	channelID int

	cb      vim.Buffer
	cw      vim.Window
	buffers []*buffer.Buffer

	bpSign map[int]*nvim.Sign // map[breakPoint.id]*nvim.Sign
	pcSign *nvim.Sign
}

// NewDelveClient represents a delve client interface.
func NewDelveClient() *delveClient {
	return &delveClient{}
}

// setupDelveClient setup the delve client. Separate the NewDelveClient() function.
// caused by neovim-go can't call the rpc2.NewClient?
func (d *delveClient) setupDelveClient(v *vim.Vim) error {
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

func (d *delveClient) cmdDebug(v *vim.Vim, eval debugEval) {
	d.debug(v, eval)
}

func (d *delveClient) debug(v *vim.Vim, eval debugEval) error {
	rootDir := context.FindVcsRoot(eval.Dir)
	srcPath := filepath.Join(os.Getenv("GOPATH"), "src") + string(filepath.Separator)
	path := filepath.Clean(strings.TrimPrefix(rootDir, srcPath))

	p := v.NewPipeline()

	if err := d.startServer("debug", path); err != nil {
		nvim.ErrorWrap(v, err)
	}
	defer d.waitServer(v)

	return d.createDebugBuffer(v, p)
}

func (d *delveClient) parseArgs(v *vim.Vim, args []string, eval createBreakpointEval) (*delveapi.Breakpoint, error) {
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

func (d *delveClient) cmdCreateBreakpoint(v *vim.Vim, args []string, eval createBreakpointEval) {
	go d.createBreakpoint(v, args, eval)
}

func (d *delveClient) createBreakpoint(v *vim.Vim, args []string, eval createBreakpointEval) error {
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

	if err := d.printLogs(v, "break "+bp.FunctionName, []byte{}); err != nil {
		return nvim.ErrorWrap(v, err)
	}

	printDebug("createBreakpoint(bp)", bp)
	return nil
}

// breakpointEval represent a breakpoint commands Eval args.
type continueEval struct {
	Dir string `msgpack:",array"`
}

func (d *delveClient) cmdContinue(v *vim.Vim, eval continueEval) {
	go d.cont(v, eval)
}

// cont sends the 'continue' signals to the delve headless server over the client use json-rpc2 protocol.
func (d *delveClient) cont(v *vim.Vim, eval continueEval) error {
	stateCh := d.client.Continue()
	state := <-stateCh

	if state == nil || state.Exited {
		return nvim.ErrorWrap(v, errors.Annotate(state.Err, pkgDelve))
	}

	cThread := state.CurrentThread
	cStacks, err := d.client.Stacktrace(cThread.GoroutineID, 1, &delveapi.LoadConfig{FollowPointers: true}) // []delveapi.Stackframe
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	if err := d.printStacktrace(v, eval.Dir, cThread, cStacks); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := d.pcSign.Place(v, cThread.ID, cThread.Line, cThread.File, true); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := v.SetWindowCursor(d.cw, [2]int{cThread.Line, 0}); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := v.Command("silent normal zz"); err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	// debug
	printDebug("stacks", cStacks)
	// printDebug("state.CurrentThread", state.CurrentThread)
	printDebug("state.SelectedGoroutine", state.SelectedGoroutine)
	// printDebug("state.Threads", state.Threads)

	return d.printLogs(v, "continue", []byte{})
}

// breakpointEval represent a breakpoint commands Eval args.
type nextEval struct {
	Dir string `msgpack:",array"`
}

func (d *delveClient) cmdNext(v *vim.Vim, eval nextEval) {
	go d.next(v, eval)
}

func (d *delveClient) next(v *vim.Vim, eval nextEval) error {
	state, err := d.client.Next()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	cThread := state.CurrentThread
	cStacks, err := d.client.Stacktrace(cThread.GoroutineID, 1, &delveapi.LoadConfig{FollowPointers: true}) // []delveapi.Stackframe
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve))
	}

	if err := d.printStacktrace(v, eval.Dir, cThread, cStacks); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := d.pcSign.Place(v, cThread.ID, cThread.Line, cThread.File, true); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := v.SetWindowCursor(d.cw, [2]int{cThread.Line, 0}); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := v.Command("silent normal zz"); err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	msg := []byte(fmt.Sprintf("%s() %s:%d  goroutine(%d) (PC: %d)", cThread.Function.Name, cThread.File, cThread.Line, cThread.GoroutineID, cThread.PC))
	return d.printLogs(v, "next", msg)
}

func (d *delveClient) cmdRestart(v *vim.Vim) {
	go d.restart(v)
}

func (d *delveClient) restart(v *vim.Vim) error {
	err := d.client.Restart()
	if err != nil {
		return nvim.ErrorWrap(v, errors.Annotate(err, pkgDelve+"restart"))
	}

	d.processPid = d.client.ProcessPid()
	return d.printLogs(v, "restart", []byte(fmt.Sprintf("Process restarted with PID %d", d.processPid)))
}

func (d *delveClient) printStacktrace(v *vim.Vim, cwd string, cThread *delveapi.Thread, cStacks []delveapi.Stackframe) error {
	v.SetBufferOption(d.buffers[2].Buffer, "modifiable", true)
	v.SetBufferOption(d.buffers[3].Buffer, "modifiable", true)
	defer v.SetBufferOption(d.buffers[2].Buffer, "modifiable", false)
	defer v.SetBufferOption(d.buffers[3].Buffer, "modifiable", false)

	var locals []byte
	stacks := []byte("\u25BC " + cThread.Function.Name + "\n")
	for _, s := range cStacks {
		if strings.HasPrefix(s.File, cwd+string(filepath.Separator)) {
			s.File = strings.TrimPrefix(s.File, cwd+string(filepath.Separator))
		}
		stacks = append(stacks, []byte(fmt.Sprintf("\t\t%s\t%s:%d\n", s.Function.Name, s.File, s.Line))...)

		for _, l := range s.Locals {
			locals = append(locals, []byte(
				fmt.Sprintf("\u25B6 %s\n\t\taddr=%d onlyAddr=%t type=%q realType=%q kind=%d value=%s len=%d cap=%d unreadable=%q\n",
					l.Name,
					l.Addr,
					l.OnlyAddr,
					l.Type,
					l.RealType,
					l.Kind,
					l.Value,
					l.Len,
					l.Cap,
					l.Unreadable))...)
		}
	}

	if err := v.SetBufferLines(d.buffers[2].Buffer, 0, -1, true, bytes.Split(stacks, []byte{'\n'})); err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	if err := v.SetBufferLines(d.buffers[3].Buffer, 0, -1, true, bytes.Split(locals, []byte{'\n'})); err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	return nil
}

func (d *delveClient) printLogs(v *vim.Vim, cmd string, message []byte) error {
	v.SetBufferOption(d.buffers[0].Buffer, "modifiable", true)
	defer v.SetBufferOption(d.buffers[0].Buffer, "modifiable", false)

	lcount, err := v.BufferLineCount(d.buffers[0].Buffer)
	if err != nil {
		return errors.Annotate(err, pkgDelve)
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

	if err := v.SetBufferLines(d.buffers[0].Buffer, lcount, -1, true, bytes.Split(msg, []byte{'\n'})); err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	afterBuf, err := v.BufferLines(d.buffers[0].Buffer, 0, -1, true)
	if err != nil {
		return errors.Annotate(err, pkgDelve)
	}

	return v.SetWindowCursor(d.buffers[0].Window, [2]int{len(afterBuf), 7})
}

func (d *delveClient) cmdState(v *vim.Vim) {
	go d.state(v)
}

func (d *delveClient) state(v *vim.Vim) error {
	state, err := d.client.GetState()
	if err != nil {
		return errors.Annotate(err, pkgDelve)
	}
	printDebug("state: %+v\n", state)
	return nil
}

func (d *delveClient) cmdStdin(v *vim.Vim) {
	go d.stdin(v)
}

// stdin sends the users input command to the internal delve terminal.
// vim input() function args is
//  input({prompt} [, {text} [, {completion}]])
// More information of input() funciton and word completion are
//  :help input()
//  :help command-completion-custom
func (d *delveClient) stdin(v *vim.Vim) error {
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

	// Output results to logs buffer.
	return d.printLogs(v, stdin.(string), out)
}

func (d *delveClient) ListFunctions(v *vim.Vim) ([]string, error) {
	funcs, err := d.client.ListFunctions("main")
	if err != nil {
		return []string{}, err
	}

	return funcs, nil
}

func (d *delveClient) readServerStdout(v *vim.Vim, cmd, args string) error {
	command := cmd + " " + args

	// Output results to logs buffer.
	return d.printLogs(v, command, d.serverOut.Bytes())
}

func (d *delveClient) readServerStderr(v *vim.Vim, cmd, args string) error {
	command := cmd + " " + args

	// Output results to logs buffer.
	return d.printLogs(v, command, d.serverErr.Bytes())
}

func printDebug(prefix string, data interface{}) error {
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("PrintDebug: %s\n%s", prefix, data)
	}
	log.Printf("PrintDebug: %s\n%s", prefix, d)

	return nil
}
