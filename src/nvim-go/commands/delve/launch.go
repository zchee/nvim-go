package delve

import (
	"bytes"
	"fmt"
	"nvim-go/nvim"
	"os/exec"

	delveapi "github.com/derekparker/delve/service/api"
	delverpc2 "github.com/derekparker/delve/service/rpc2"
	delveterminal "github.com/derekparker/delve/terminal"
	"github.com/garyburd/neovim-go/vim"
)

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
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>c    :<C-u>call rpcrequest(%d, 'DlvContinue')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>n    :<C-u>call rpcrequest(%d, 'DlvNext')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>s    :<C-u>call rpcrequest(%d, 'DlvStep')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>r    :<C-u>call rpcrequest(%d, 'DlvRestart')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>i    :<C-u>call rpcrequest(%d, 'DelveStdin')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent><CR> :<C-u>call rpcrequest(%d, 'DelveLastCmd')<CR>", channelId))
	p.Command(fmt.Sprintf("nnoremap <buffer><silent>q    :<C-u>call rpcrequest(%d, 'DlvDetach')<CR>", channelId))

	return p.Wait()
}
