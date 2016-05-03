package nvim

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
)

var (
	tbuffer vim.Buffer
	twindow vim.Window
)

// Terminal configure of open the terminal.
type Terminal struct {
	v    *vim.Vim
	cmd  []string
	mode string
	// Dir specifies the working directory of the command on terminal.
	Dir string
	// Width split window width for open the terminal window.
	Width int64
	// Height split window height for open the terminal window.
	Height int64
}

// NewTerminal return the initialize Neovim terminal config.
func NewTerminal(vim *vim.Vim, command []string, mode string) *Terminal {
	return &Terminal{
		v:    vim,
		cmd:  command,
		mode: mode,
	}
}

// Run runs the command in the terminal buffer.
func (t *Terminal) Run() error {
	if t.Dir != "" {
		defer chdir(t.v, t.Dir)()
	}

	var (
		b      vim.Buffer
		w      vim.Window
		pos    = config.TerminalPosition
		height = config.TerminalHeight
		width  = config.TerminalWidth
	)

	// Creates a new pipeline
	p := t.v.NewPipeline()
	p.CurrentBuffer(&b)
	p.CurrentWindow(&w)
	if err := p.Wait(); err != nil {
		return err
	}

	if twindow != 0 {
		p.SetCurrentWindow(twindow)
		p.SetBufferOption(tbuffer, "modified", false)
		p.Call("termopen", nil, strings.Join(t.cmd, " "))
		p.SetBufferOption(tbuffer, "modified", true)
	} else {
		// Set split window position. (defalut: botright)
		vcmd := pos + " "

		t.Height = height
		t.Width = width

		switch {
		case t.Height != int64(0) && t.mode == "split":
			vcmd += strconv.FormatInt(t.Height, 10)
		case t.Width != int64(0) && t.mode == "vsplit":
			vcmd += strconv.FormatInt(t.Width, 10)
		case strings.Index(t.mode, "split") == -1:
			return fmt.Errorf("%s mode is not supported", t.mode)
		}

		// Create terminal buffer and spawn command.
		vcmd += t.mode + " | terminal " + strings.Join(t.cmd, " ")
		p.Command(vcmd)

		// Get terminal buffer and windows information.
		p.CurrentBuffer(&tbuffer)
		p.CurrentWindow(&twindow)
		if err := p.Wait(); err != nil {
			return err
		}

		// Workaround for "autocmd BufEnter term://* startinsert"
		if config.TerminalStartInsert {
			p.Command("stopinsert")
		}

		p.SetBufferOption(tbuffer, "filetype", "terminal")
		p.SetBufferOption(tbuffer, "buftype", "nofile")
		p.SetBufferOption(tbuffer, "bufhidden", "delete")
		p.SetBufferOption(tbuffer, "buflisted", false)
		p.SetBufferOption(tbuffer, "swapfile", false)

		p.SetWindowOption(twindow, "list", false)
		p.SetWindowOption(twindow, "number", false)
		p.SetWindowOption(twindow, "relativenumber", false)
		p.SetWindowOption(twindow, "winfixheight", true)

		p.Command("autocmd! * <buffer>")
		p.Command("autocmd BufEnter <buffer> startinsert")
	}

	// Set buffer name, filetype and options
	p.SetBufferName(tbuffer, "__GO_TERMINAL__")

	// Refocus coding buffer and stop insert mode
	p.SetCurrentWindow(w)
	p.Command("stopinsert")

	return p.Wait()
}

// chdir changes vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func chdir(v *vim.Vim, dir string) func() {
	var (
		m   sync.Mutex
		cwd interface{}
	)
	m.Lock()
	if err := v.Eval("getcwd()", &cwd); err != nil {
		Echoerr(v, "GoTerminal: %v", err)
	}
	v.ChangeDirectory(dir)
	return func() {
		v.ChangeDirectory(cwd.(string))
		m.Unlock()
	}
}
