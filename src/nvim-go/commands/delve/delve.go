// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package delve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"nvim-go/nvim"

	delveapi "github.com/derekparker/delve/service/api"
	delverpc2 "github.com/derekparker/delve/service/rpc2"
	delveterminal "github.com/derekparker/delve/terminal"
	"github.com/garyburd/neovim-go/vim"
)

const addr = "localhost:41222" // d:4 l:12 v:22

var (
	delve  *DelveClient
	server *exec.Cmd

	stdout, stderr bytes.Buffer

	channelId   int
	baseTabpage vim.Tabpage

	// TODO(zchee): More elegant way.
	src    = &bufferInfo{}
	logs   = &bufferInfo{}
	breaks = &bufferInfo{}
	stacks = &bufferInfo{}
	locals = &bufferInfo{}
)

type bufferInfo struct {
	buffer vim.Buffer
	window vim.Window

	bufnr     interface{}
	linecount int
	name      string
}

// DelveClient represents a delve debugger interface and buffer information.
type DelveClient struct {
	client   *delverpc2.RPCClient
	terminal *delveterminal.Term
	debugger *delveterminal.Commands

	addr    string
	procPid int

	buffers     map[vim.Buffer]*bufferInfo
	breakpoints map[int]*delveapi.Breakpoint
	bpSign      map[string]*nvim.Sign
	pcSign      *nvim.Sign
	lastBpId    int
}

// NewDelveClient represents a delve client interface.
func NewDelveClient(addr string) *DelveClient {
	// TODO(zchee): custimizable listen address. Now use constant port.
	// delve can remote debugging of another PC over the http?
	// and can debug any binary in the Docker container?
	return &DelveClient{
		addr: addr,
	}
}

// startServer starts the delve headless server and hijacked stdout & stderr.
func delveStartServer(v *vim.Vim, args []string, eval cmdDelveEval) error {
	bin, err := exec.LookPath("astdump")
	if err != nil {
		return err
	}

	serverArgs := []string{"exec", bin, "--headless=true", "--accept-multiclient=true", "--api-version=2", "--log", "--listen=" + addr}
	server = exec.Command("dlv", serverArgs...)

	server.Stdout = &stdout
	server.Stderr = &stderr

	err = server.Run()
	if err != nil {
		return err
	}

	return nil
}

// dlvStartClient starts the delve client use json-rpc2 protocol.
func delveStartClient(v *vim.Vim, eval cmdDelveEval) error {
	if server == nil {
		return nvim.EchohlErr(v, "Delve", "dlv headless server not running")
	}

	delve = NewDelveClient(addr)
	delve.client = delverpc2.NewClient(addr)
	delve.procPid = delve.client.ProcessPid()
	delve.buffers = make(map[vim.Buffer]*bufferInfo, 5)

	delve.terminal = delveterminal.New(delve.client, nil)
	delve.debugger = delveterminal.DebugCommands(delve.client)

	channelId, _ = v.ChannelID()
	baseTabpage, _ = v.CurrentTabpage()

	p := v.NewPipeline()
	newBuffer(p, "source", "0tab", 0, "new", src)
	p.Command("runtime! syntax/go.vim")
	if err := p.Wait(); err != nil {
		return err
	}

	// Define sign for breakpoint hit line.
	// TODO(zchee): Custumizable sign text and highlight group.
	var width, height int
	var err error
	delve.pcSign, err = nvim.NewSign(v, "delve_pc", "->", "String", "Search")
	delve.bpSign = make(map[string]*nvim.Sign)
	p.Command("sign define delve_bp text=B> texthl=Type")
	p.WindowWidth(src.window, &width)
	p.WindowHeight(src.window, &height)
	if err := p.Wait(); err != nil {
		return err
	}

	newBuffer(p, "stacktrace", "belowright", (width * 2 / 5), "vsplit", stacks)
	newBuffer(p, "breakpoint", "belowright", (height * 1 / 3), "split", breaks)
	newBuffer(p, "locals", "belowright", (height * 1 / 3), "split", locals)
	p.SetCurrentWindow(src.window)
	if err := p.Wait(); err != nil {
		return err
	}
	newBuffer(p, "logs", "belowright", (height * 1 / 3), "split", logs)
	if err := p.Wait(); err != nil {
		return err
	}

	// Gets the default "unrecovered-panic" breakpoint
	delve.breakpoints = make(map[int]*delveapi.Breakpoint)

	unrecovered, err := delve.client.GetBreakpoint(-1)
	if err != nil {
		return nvim.EchohlErr(v, "Delve", err)
	}

	delve.breakpoints[-1] = unrecovered
	delve.bpSign[unrecovered.File], err = nvim.NewSign(v, "delve_bp", "B>", "Type", "")
	if err != nil {
		return nvim.EchohlErr(v, "Delve", err)
	}

	ubp := formatBreakpoint(unrecovered)
	breaks.linecount = printBufferPipe(p, breaks.buffer, false, bytes.Split(ubp, []byte{'\n'}))

	// TODO(zchee): Workaround for "API server listening at..." first server stdout.
	stdout.Reset()

	return p.Wait()
}

func newBuffer(p *vim.Pipeline, name string, mode string, size int, split string, buf *bufferInfo) error {
	buf.name = name
	p.Command(fmt.Sprintf("silent %s %d%s [delve] %s", mode, size, split, buf.name))
	if err := p.Wait(); err != nil {
		return err
	}

	p.CurrentBuffer(&buf.buffer)
	p.CurrentWindow(&buf.window)
	if err := p.Wait(); err != nil {
		return err
	}

	delve.buffers[buf.buffer] = buf

	p.Eval("bufnr('%')", &buf.bufnr)
	p.SetBufferOption(buf.buffer, "filetype", "delve")
	p.SetBufferOption(buf.buffer, "buftype", "nofile")
	p.SetBufferOption(buf.buffer, "bufhidden", "delete")
	p.SetBufferOption(buf.buffer, "buflisted", false)
	p.SetBufferOption(buf.buffer, "swapfile", false)
	p.SetWindowOption(buf.window, "winfixheight", true)
	if buf.name != "source" {
		p.SetWindowOption(buf.window, "list", false)
		p.SetWindowOption(buf.window, "number", false)
		p.SetWindowOption(buf.window, "relativenumber", false)
	}
	// modifiable lock.
	p.SetBufferOption(buf.buffer, "modifiable", false)
	if err := p.Wait(); err != nil {
		return err
	}
	// TODO(zchee): Why can't use p.SetBufferOption?
	p.Call("setbufvar", nil, buf.bufnr.(int64), "&colorcolumn", "")

	// TODO(zchee): Move to <Plug> mappnig when releases.
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>c :<C-u>call rpcrequest(%d, 'DlvContinue')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>n :<C-u>call rpcrequest(%d, 'DlvNext')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>r :<C-u>call rpcrequest(%d, 'DlvRestart')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>q :<C-u>call rpcrequest(%d, 'DlvDetach')<CR>", channelId))

	return p.Wait()
}

// delveCommand sends the users input delve subcommand and arguments to the internal launched delve vertual terminal.
func delveCommand(v *vim.Vim, args []string) error {
	// Create the connected pair of *os.Files and replace os.Stdout.
	// delve terminal return to stdout only.
	r, w, _ := os.Pipe() // *os.File
	saveStdout := os.Stdout
	os.Stdout = w

	// First command arguments is delve subcommand.
	// Splits the after arguments with whitespace.
	err := delve.debugger.Call(args[0], strings.Join(args[1:], " "), delve.terminal)
	if err != nil {
		return err
	}

	// Close the w file and restore os.Stdout to original.
	w.Close()
	os.Stdout = saveStdout

	if err := printLogs(v, []byte(strings.Join(args, " ")), true); err != nil {
		return err
	}
	// Read all the lines of r file and output results to logs buffer.
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := printLogs(v, out, false); err != nil {
		return err
	}

	return nil
}

// ByID sorts breakpoints by ID.
type ByID []*delveapi.Breakpoint

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

func delveSetBreakpoint(v *vim.Vim, args []string) error {
	var bpName string

	switch len(args) {
	case 0:
		return nvim.EchohlErr(v, "Delve", "Invalid argument")
	case 1:
		// TODO(zchee): more elegant way
		bpslice := strings.Split(args[0], ".")
		bpslice[1] = fmt.Sprintf("%s%s", strings.ToUpper(bpslice[1][:1]), bpslice[1][1:])
		bpName = strings.Join(bpslice, "")
	case 2:
		bpName = args[1]
	default:
		return nvim.EchohlErr(v, "Delve", "Too many arguments")
	}

	newbp, err := delve.client.CreateBreakpoint(&delveapi.Breakpoint{
		FunctionName: args[0],
		Name:         bpName,
		Tracepoint:   true,
	}) // *api.Breakpoint
	if err != nil {
		return nvim.EchohlErr(v, "Delve", err)
	}
	delve.breakpoints[newbp.ID] = newbp
	if delve.bpSign[newbp.File] == nil {
		delve.bpSign[newbp.File], err = nvim.NewSign(v, "delve_bp", "B>", "Type", "")
		if err != nil {
			return nvim.EchohlErr(v, "Delve", err)
		}
	}

	// Breakpoint 1 at 0x2053 for main.main() /Users/zchee/go/src/github.com/zchee/go-sandbox/astdump/astdump.go:19 (1)
	bp := formatBreakpoint(newbp)
	if breaks.linecount, err = printBuffer(v, breaks.buffer, true, bytes.Split(bp, []byte{'\n'})); err != nil {
		return nvim.EchohlErr(v, "Delve", err)
	}
	if err := v.SetWindowCursor(breaks.window, [2]int{breaks.linecount, 0}); err != nil {
		return nvim.EchohlErr(v, "Delve", err)
	}

	return printLogs(v, []byte(fmt.Sprintf("Breakpoint %d set at %#x for %s() %s:%d", newbp.ID, newbp.Addr, newbp.FunctionName, newbp.File, newbp.Line)), true)
}

func delveFunctionList(v *vim.Vim) ([]string, error) {
	funcs, err := delve.client.ListFunctions("main")
	if err != nil {
		return []string{}, nil
	}

	return funcs, nil
}

// parseThread parses the delve Thread information and print the each result
// to the corresponding buffer.
//
// delve original stdout output sample:
//  // continue
//  > main.main() /Users/zchee/go/src/github.com/zchee/golist/golist.go:29 (hits goroutine(1):1 total:1) (PC: 0x20eb)
//  // next
//  > runtime.main() /usr/local/go/src/runtime/proc.go:182 (PC: 0x26e2a)
func parseThread(v *vim.Vim, thread *delveapi.Thread) error {
	if thread != nil {
		p := v.NewPipeline()
		if src.name != thread.File {
			byt, err := ioutil.ReadFile(thread.File)
			if err != nil {
				return err
			}
			src.name = thread.File

			p.SetBufferName(src.buffer, thread.File)
			if _ = printBufferPipe(p, src.buffer, false, bytes.Split(byt, []byte{'\n'})); err != nil {
				return err
			}
			delve.bpSign[thread.File].UnplaceAll(p, thread.File)
			for _, bp := range delve.breakpoints {
				if bp.File == thread.File {
					delve.bpSign[thread.File].Place(p, bp.ID, bp.Line, thread.File, false)
				}
			}
		}

		delve.pcSign.Place(p, thread.ID, thread.Line, thread.File, true)
		p.SetWindowCursor(src.window, [2]int{thread.Line, 0})
		err := p.Wait()
		if err != nil {
			return err
		}

		if stdout.Len() != 0 {
			printLogs(v, stdout.Bytes(), false)
			defer stdout.Reset()
		}

		funcName := fmt.Sprintf("%s() ", thread.Function.Name)
		file := fmt.Sprintf("%s", thread.File)
		line := fmt.Sprintf(":%d ", thread.Line)
		goroutine := fmt.Sprintf("goroutine(%d) ", thread.GoroutineID)
		pc := fmt.Sprintf("(PC: %#x)", thread.PC)

		printLogs(v, ([]byte("> " + funcName + file + line + goroutine + pc)), false)
	}
	return nil
}

// delveContinue sends the 'continue' signals to the delve headless server over the client use json-rpc2 protocol.
func delveContinue(v *vim.Vim) error {
	stateCh := delve.client.Continue()
	state := <-stateCh

	if state == nil || state.Exited {
		p := v.NewPipeline()
		delve.pcSign.UnplaceAllPc(p)
		return nvim.EchohlErr(v, "Delve", fmt.Sprintf("%s", state.Err))
	}

	if err := printLogs(v, []byte("continue"), true); err != nil {
		return err
	}

	if err := parseThread(v, state.CurrentThread); err != nil {
		return err
	}
	if err := updateBreakpoint(v); err != nil {
		return err
	}

	return nil
}

// delveNext sends the 'next' signals to the delve headless server over the client use json-rpc2 protocol.
func delveNext(v *vim.Vim) error {
	state, err := delve.client.Next()
	if err != nil {
		p := v.NewPipeline()
		delve.pcSign.UnplaceAllPc(p)
		return nvim.EchohlErr(v, "Delve", fmt.Sprintf("%s", err))
	}

	if err := printLogs(v, []byte("next"), true); err != nil {
		return err
	}

	if err := parseThread(v, state.CurrentThread); err != nil {
		return err
	}
	if err := updateBreakpoint(v); err != nil {
		return err
	}

	return nil
}

func updateBreakpoint(v *vim.Vim) error {
	breakpoint, err := delve.client.ListBreakpoints()
	if err != nil {
		return err
	}
	sort.Sort(ByID(breakpoint))

	var bplines []byte
	for _, bp := range breakpoint {
		if delve.breakpoints[bp.ID].TotalHitCount != bp.TotalHitCount {
			delve.breakpoints[bp.ID].TotalHitCount = bp.TotalHitCount
			delve.breakpoints[bp.ID].HitCount = bp.HitCount
		}
		bufbp := formatBreakpoint(bp)
		bplines = append(bplines, bufbp...)
		bplines = append(bplines, byte('\n'))
	}

	if breaks.linecount, err = printBuffer(v, breaks.buffer, false, bytes.Split(bplines, []byte{'\n'})); err != nil {
		return err
	}

	return v.SetWindowCursor(breaks.window, [2]int{breaks.linecount, 0})
}

func formatBreakpoint(breakpoint *delveapi.Breakpoint) []byte {
	bp := bytes.NewBufferString(
		fmt.Sprintf("%2d: PC=%#x func=%s() File=%s:%d",
			breakpoint.ID,
			breakpoint.Addr,
			breakpoint.FunctionName,
			breakpoint.File,
			breakpoint.Line))

	return bp.Bytes()
}

func printLogs(v *vim.Vim, message []byte, prefix bool) error {
	var msg []byte
	var err error

	if prefix {
		msg = []byte("(dlv) ")
	}

	msg = append(msg, bytes.TrimSpace(message)...)
	logs.linecount, err = printBuffer(v, logs.buffer, true, bytes.Split(msg, []byte{'\n'}))
	if err != nil {
		return err
	}

	return v.SetWindowCursor(logs.window, [2]int{logs.linecount, 0})
}

func printBuffer(v *vim.Vim, b vim.Buffer, append bool, data [][]byte) (int, error) {
	var start int

	// Gets the buffer line count if append is true.
	if append {
		var err error
		start, err = v.BufferLineCount(b)
		if err != nil {
			return 0, err
		}
	}

	// Chceck the target buffer whether empty if line count is 1.
	if start == 1 {
		buf, err := v.BufferLines(b, 0, -1, true)
		if err != nil {
			return 0, err
		}
		// buf[0] is target buffer's first line []byte slice.
		if len(buf[0]) == 0 {
			start = 0
		}
	}

	v.SetBufferOption(b, "modifiable", true)
	defer v.SetBufferOption(b, "modifiable", false)

	return start + len(data), v.SetBufferLines(b, start, -1, true, data)
}

func printBufferPipe(p *vim.Pipeline, b vim.Buffer, append bool, data [][]byte) int {
	var start int

	// Gets the buffer line count if append is true.
	if append {
		p.BufferLineCount(b, &start)
	}

	// Chceck the target buffer whether empty if line count is 1.
	if start == 1 {
		var buf [][]byte
		p.BufferLines(b, 0, -1, true, &buf)
		// buf[0] is target buffer's first line []byte slice.
		if len(buf[0]) == 0 {
			start = 0
		}
	}

	p.SetBufferOption(b, "modifiable", true)
	defer p.SetBufferOption(b, "modifiable", false)

	p.SetBufferLines(b, start, -1, true, data)
	return start + len(data)
}

func delveDisassemble(v *vim.Vim) error {
	// delve.c.DisassemblePC()
	return nil
}

func delveRestart(v *vim.Vim) error {
	err := delve.client.Restart()
	if err != nil {
		return err
	}

	return printLogs(v, []byte("restart"), true)
}

func delveDetach(v *vim.Vim) error {
	defer delveKill()
	if delve.procPid == 0 {
		return nil
	}

	if delve.buffers != nil {
		p := v.NewPipeline()
		p.SetCurrentTabpage(baseTabpage)
		for _, buf := range delve.buffers {
			p.Command(fmt.Sprintf("bdelete %d", buf.bufnr))
		}

		if err := p.Wait(); err != nil {
			return err
		}
	}
	err := delve.client.Detach(true)
	if err != nil {
		return err
	}
	log.Println("Detached delve client")

	return nil
}

func delveKill() error {
	if server != nil {
		err := server.Process.Kill()
		if err != nil {
			return err
		}
		log.Println("Killed delve server")
	}

	return nil
}

func delvePrintDebug(prefix string, data interface{}) error {
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Println("PrintDebug: ", prefix, "\n", data)
	}
	log.Println("PrintDebug: ", prefix, "\n", string(d))

	return nil
}
