package terminal

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"nvim-go/config"
	"nvim-go/nvim"

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

	Filetype  string
	Buftype   string
	Bufhidden string
	Buflisted bool
	Swapfile  bool

	List           bool
	Number         bool
	Relativenumber bool
	Winfixheight   bool
}

// NewTerminal return the initialize Neovim terminal config.
func NewTerminal(vim *vim.Vim, command []string, mode string) *Terminal {
	return &Terminal{
		v:    vim,
		cmd:  command,
		mode: mode,

		Filetype:  "terminal",
		Buftype:   "nofile",
		Bufhidden: "delete",
		Buflisted: false,
		Swapfile:  false,

		List:           false,
		Number:         false,
		Relativenumber: false,
		Winfixheight:   true,
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

		p.SetBufferOption(tbuffer, "filetype", t.Filetype)
		p.SetBufferOption(tbuffer, "buftype", t.Buftype)
		p.SetBufferOption(tbuffer, "bufhidden", t.Bufhidden)
		p.SetBufferOption(tbuffer, "buflisted", t.Buflisted)
		p.SetBufferOption(tbuffer, "swapfile", t.Swapfile)

		p.SetWindowOption(twindow, "list", t.List)
		p.SetWindowOption(twindow, "number", t.Number)
		p.SetWindowOption(twindow, "relativenumber", t.Relativenumber)
		p.SetWindowOption(twindow, "winfixheight", t.Winfixheight)

		// Cleanup cursor highlighting
		// TODO(zchee): Can use p.AddBufferHighlight?
		p.Command("hi TermCursor gui=NONE guifg=NONE guibg=NONE")
		p.Command("hi! TermCursorNC gui=NONE guifg=NONE guibg=NONE")

		// Cleanup autocmd for terminal buffer
		// The following autocmd is defined only in the terminal buffer local
		p.Command("autocmd! * <buffer>")
		// Set autocmd of automatically insert mode
		p.Command("autocmd WinEnter <buffer> startinsert")
		// Set autoclose buffer if the current buffer is only terminal
		// TODO(zchee): convert to rpc way
		p.Command("autocmd WinEnter <buffer> if winnr('$') == 1 | quit | endif")
	}

	// Set buffer name, filetype and options
	p.SetBufferName(tbuffer, "__GO_TERMINAL__")

	// Refocus coding buffer and stop insert mode
	p.SetCurrentWindow(w)
	p.Command("stopinsert")

	return p.Wait()
}

func Command(v *vim.Vim, w vim.Window, command string) error {
	defer switchFocus(v, w)()

	var b vim.Buffer
	p := v.NewPipeline()
	p.CurrentBuffer(&b)
	p.SetBufferOption(b, "modified", false)
	p.Wait()

	p.FeedKeys("i"+command+"\r", "n", true)
	p.SetBufferOption(b, "modified", true)
	log.Printf("command: %+v\n", command)
	p.Command("stopinsert")

	return p.Wait()
}

// TODO(zchee): flashing when switch the window.
func switchFocus(v *vim.Vim, w vim.Window) func() {
	var (
		m  sync.Mutex
		cw vim.Window
	)
	m.Lock()

	p := v.NewPipeline()
	p.CurrentWindow(&cw)

	if err := p.Wait(); err != nil {
		nvim.Echoerr(v, "GoTerminal: %v", err)
	}

	p.SetCurrentWindow(w)

	return func() {
		v.SetCurrentWindow(cw)
		m.Unlock()
	}
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
		nvim.Echoerr(v, "GoTerminal: %v", err)
	}
	v.ChangeDirectory(dir)
	return func() {
		v.ChangeDirectory(cwd.(string))
		m.Unlock()
	}
}
