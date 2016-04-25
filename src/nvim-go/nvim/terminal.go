package nvim

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
)

type Terminal struct {
	v      *vim.Vim
	cmd    []string
	mode   string
	pos    string
	Width  int64
	Height int64
}

func NewTerminal(vim *vim.Vim, command []string) *Terminal {
	return &Terminal{
		v:    vim,
		cmd:  command,
		mode: config.TerminalMode,
		pos:  config.TerminalPosition,
	}
}

func (t *Terminal) Run() error {
	var (
		b      vim.Buffer
		w      vim.Window
		tb     vim.Buffer
		tw     vim.Window
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

	// Set split window position. (defalut: botright)
	vcmd := t.pos + " "

	switch {
	case height != int64(0) && t.mode == "split":
		vcmd = strconv.FormatInt(height, 10)
	case width != int64(0) && t.mode == "vsplit":
		vcmd = strconv.FormatInt(width, 10)
	case strings.Index(t.mode, "split") == -1:
		return errors.New(fmt.Sprintf("%s mode is not supported", t.mode))
	}

	// Create terminal buffer and spawn command.
	vcmd += t.mode + " | terminal " + strings.Join(t.cmd, " ")
	p.Command(vcmd)

	// Get terminal buffer and windows information.
	p.CurrentBuffer(&tb)
	p.CurrentWindow(&tw)
	if err := p.Wait(); err != nil {
		return err
	}

	// Set buffer name, filetype and options
	p.SetBufferName(tb, "__GO_TERMINAL__")

	p.SetBufferOption(tb, "filetype", "goterm")
	p.SetBufferOption(tb, "bufhidden", "delete")
	p.SetBufferOption(tb, "buflisted", false)
	p.SetBufferOption(tb, "swapfile", false)

	p.SetWindowOption(tw, "list", false)
	p.SetWindowOption(tw, "number", false)
	p.SetWindowOption(tw, "winfixheight", true)

	// Refocus coding buffer
	p.SetCurrentWindow(w)
	// Workaround for "autocmd BufEnter term://* startinsert"
	if config.TerminalStartInsert == int64(0) {
		p.Command("stopinsert")
	}

	return p.Wait()
}

func (t *Terminal) Open() error {
	return nil
}
