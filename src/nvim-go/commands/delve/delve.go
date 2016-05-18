package delve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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
	"github.com/pkg/errors"
)

const (
	addr = "localhost:41222" // d:4 l:12 v:22

	pkgDelve = "Delve"
)

func init() {
	d := NewDelveClient()
	// Launch
	plugin.HandleCommand("DlvDebug", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p:h')]"}, d.cmdDebug)
	// Breakpoint
	plugin.HandleCommand("DlvSetBreakpoint", &plugin.CommandOptions{Eval: "[expand('%:p')]"}, d.cmdBreakpoint)
	// Stdin
	plugin.HandleCommand("DlvStdin", &plugin.CommandOptions{}, d.stdin)
	// State
	plugin.HandleCommand("DlvState", &plugin.CommandOptions{}, d.cmdState)
	// Detach
	plugin.HandleCommand("DlvDetach", &plugin.CommandOptions{}, d.cmdDetach)

	// autocmd VimLeavePre
	plugin.HandleAutocmd("VimLeavePre", &plugin.AutocmdOptions{Group: "nvim-go", Pattern: "*.go"}, d.cmdDetach)
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

	bpSign map[int]*nvim.Sign
	pcSign *nvim.Sign
}

// NewDelveClient represents a delve client interface.
func NewDelveClient() *delveClient {
	return &delveClient{}
}

// SetupDelveClient setup the delve client.
// It's separate the NewDelveClient() function.
// caused by neovim-go can't call the rpc2.NewClient?
func (d *delveClient) SetupDelveClient(v *vim.Vim) error {
	var err error

	d.client = delverpc2.NewClient(addr)           // *rpc2.RPCClient
	d.term = delveterm.New(d.client, nil)          // *terminal.Term
	d.debugger = delveterm.DebugCommands(d.client) // *terminal.Commands
	d.channelID, err = v.ChannelID()               // int
	d.processPid = d.client.ProcessPid()           // int
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	return nil
}

// startServer starts the delve headless server and hijacked stdout & stderr.
func (d *delveClient) startServer(cmd, path string) error {
	dlvBin, err := exec.LookPath("dlv")
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	// TODO(zchee): costomizable build flag
	args := []string{cmd, path, "--headless=true", "--accept-multiclient=true", "--api-version=2", "--log", "--listen=" + addr}
	d.server = exec.Command(dlvBin, args...)

	d.server.Stdout = &d.serverOut
	d.server.Stderr = &d.serverErr

	if err := d.server.Start(); err != nil {
		err = errors.New(d.serverOut.String())
		defer d.serverOut.Reset()
		return errors.Wrap(err, pkgDelve)
	}

	return nil
}

func (d *delveClient) waitServer(v *vim.Vim) error {
	// Waiting for dlv launch the headless server.
	// "net.Dial" is better way?
	nvim.EchoProgress(v, "Delve", "Wait for running dlv server")
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		break
	}
	if err := d.SetupDelveClient(v); err != nil {
		return nvim.EchoerrWrap(v, errors.Wrap(err, pkgDelve))
	}
	return nvim.Echomsg(v, "Delve: Launched dlv headless server")
}

func (d *delveClient) createDebugBuffer(v *vim.Vim) error {
	p := v.NewPipeline()
	p.CurrentBuffer(&d.cb)
	p.CurrentWindow(&d.cw)
	if err := p.Wait(); err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	var height, width int
	p.WindowHeight(d.cw, &height)
	p.WindowWidth(d.cw, &width)
	if err := p.Wait(); err != nil {
		return errors.Wrap(err, pkgDelve)
	}

	bufOption := d.setNvimOption("buffer")
	winOption := d.setNvimOption("window")

	d.buffers = make([]*buffer.Buffer, 3)
	for i, n := range []string{"stacktarce", "breakpoints", "locals"} {
		d.buffers[i] = buffer.NewBuffer(n)
	}
	d.buffers[0].Mode = fmt.Sprintf("belowright %d vsplit", (width * 2 / 5))
	d.buffers[1].Mode = fmt.Sprintf("belowright %d split", (height * 1 / 3))
	d.buffers[2].Mode = fmt.Sprintf("belowright %d split", (height * 1 / 3))

	for _, buf := range d.buffers {
		if err := buf.Create(v, bufOption, winOption); err != nil {
			return errors.Wrap(err, pkgDelve)
		}
	}

	p.SetCurrentWindow(d.cw)
	return p.Wait()
}

func (d *delveClient) setNvimOption(scope string) map[string]interface{} {
	options := make(map[string]interface{})

	switch scope {
	case "buffer":
		options[buffer.Bufhidden] = buffer.BufhiddenDelete
		options[buffer.Buflisted] = false
		options[buffer.Buftype] = buffer.BuftypeNofile
		options[buffer.Filetype] = buffer.FiletypeGo
		options[buffer.Swapfile] = false
	case "window":
		options[buffer.List] = false
		options[buffer.Number] = false
		options[buffer.Relativenumber] = false
		options[buffer.Winfixheight] = true
	}

	return options
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

	if err := d.startServer("debug", path); err != nil {
		return nvim.EchoerrWrap(v, err)
	}

	if err := d.createDebugBuffer(v); err != nil {
		return nvim.EchoerrWrap(v, err)
	}
	defer d.state(v)

	defer d.waitServer(v)
	return nil
}

func (d *delveClient) cmdState(v *vim.Vim) {
	go d.state(v)
}

func (d *delveClient) state(v *vim.Vim) error {
	state, err := d.client.GetState()
	if err != nil {
		return errors.Wrap(err, pkgDelve)
	}
	printDebug("state: %+v\n", state)
	return nil
}

// breakpointEval represent a breakpoint commands Eval args.
type breakpointEval struct {
	File string `msgpack:",array"`
}

func (d *delveClient) cmdBreakpoint(v *vim.Vim, eval breakpointEval) {
	go d.breakpoint(v, eval)
}

func (d *delveClient) breakpoint(v *vim.Vim, eval breakpointEval) error {
	if d.bpSign == nil {
		d.bpSign = make(map[int]*nvim.Sign)
	}

	cursor, err := v.WindowCursor(d.cw)
	if err != nil {
		return nvim.EchoerrWrap(v, errors.Wrap(err, pkgDelve))
	}

	bp, err := d.client.CreateBreakpoint(&delveapi.Breakpoint{
		File: eval.File,
		Line: cursor[0],
	}) // *delveapi.Breakpoint
	if err != nil {
		return nvim.EchoerrWrap(v, errors.Wrap(err, pkgDelve))
	}

	printDebug("setBreakpoint(bp): %+v\n", bp)
	return nil
}

func (d *delveClient) cmdStdin(v *vim.Vim) {
	go d.stdin(v)
}

// stdin sends the users input command to the internal delve terminal.
func (d *delveClient) stdin(v *vim.Vim) error {
	var stdin interface{}
	err := v.Call("input", &stdin, "(dlv) ")
	if err != nil {
		return nil
	}

	if stdin.(string) != "" {
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

		err := d.debugger.Call(cmd[0], args, d.term)
		if err != nil {
			return nvim.EchoerrWrap(v, errors.Wrap(err, pkgDelve))
		}

		// Close the w file and restore os.Stdout to original.
		w.Close()
		os.Stdout = saveStdout

		// Read all the lines of r file and output results to logs buffer.
		out, err := ioutil.ReadAll(r)
		if err != nil {
			return nvim.EchoerrWrap(v, errors.Wrap(err, pkgDelve))
		}
		nvim.EchoRaw(v, string(out))
	}

	return nil
}

func (d *delveClient) cmdDetach(v *vim.Vim) {
	go d.detach(v)
}

func (d *delveClient) detach(v *vim.Vim) error {
	defer d.kill()
	if d.processPid != 0 {
		err := d.client.Detach(true)
		if err != nil {
			return errors.Wrap(err, pkgDelve)
		}
		log.Printf("Detached delve client\n")
	}

	return nil
}

func (d *delveClient) kill() error {
	if d.server != nil {
		err := d.server.Process.Kill()
		if err != nil {
			return errors.Wrap(err, pkgDelve)
		}
		log.Printf("Killed delve server\n")
	}

	return nil
}

func printDebug(prefix string, data interface{}) error {
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Println("PrintDebug: ", prefix, "\n", data)
	}
	log.Println("PrintDebug: ", prefix, "\n", string(d))

	return nil
}
