// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

const pkgNvimSign = "nvim/sign"

var (
	BreakpointSymbol         = "\u25cf" // ●  BLACK CIRCLE                         (U+25CF)
	BreakpointSymbolLarge    = "\u2b24" // ⬤  BLACK LARGE CIRCLE                   (U+2B24)
	TracepointSymbol         = "\u25c6" // ◆  BLACK DIAMOND                        (U+25C6)
	TracepointSymbolMidium   = "\u2b25" // ⬥  BLACK DIAMOND SUIT                   (U+2B25)
	ProgramCounterSymbol     = "\u25ce" // ◎  BULLSEYE                             (U+25CE)
	ProgramCounterSymbolRing = "\u23e3" // ⏣  BENZENE RING WITH CIRCLE             (U+23e3)
	ErrorSymbol              = "\u20e0" //  ⃠  COMBINING ENCLOSING CIRCLE BACKSLASH (U+20E0)
	RestartSymbol            = "\u27f2" // ⟲  ANTICLOCKWISE GAPPED CIRCLE ARROW    (U+27F2)
)

type Sign struct {
	Name   string
	Text   string
	Texthl string
	Linehl string

	LastID   int
	LastLine int
	LastFile string
}

func NewSign(v *vim.Vim, name, text, texthl, linehl string) (*Sign, error) {
	cmd := fmt.Sprintf("sign define %s", name)
	switch {
	case text != "":
		cmd += " text=" + text
		fallthrough
	case texthl != "":
		cmd += " texthl=" + texthl
		fallthrough
	case linehl != "":
		cmd += " linehl=" + linehl
	}

	if err := v.Command(cmd); err != nil {
		return nil, err
	}

	return &Sign{
		Name:   name,
		Text:   text,
		Texthl: texthl,
		Linehl: linehl,
	}, nil
}

func (s *Sign) Place(v *vim.Vim, id, line int, file string, clearLastSign bool) error {
	if clearLastSign && s.LastID != 0 {
		place := fmt.Sprintf("sign unplace %d file=%s", s.LastID, file)
		v.Command(place)
	}

	// TODO(zchee): workaroud for "unrecovered-panic" default breakpoint.
	if id < 0 {
		id = 99
	}
	place := fmt.Sprintf("sign place %d name=%s line=%d file=%s", id, s.Name, line, file)
	if err := v.Command(place); err != nil {
		return errors.Annotate(err, pkgNvimSign)
	}
	s.LastID = id
	s.LastFile = file

	return nil
}

func (s *Sign) PlacePipe(p *vim.Pipeline, id, line int, file string, clearLastSign bool) error {
	if clearLastSign && s.LastID != 0 {
		cmd := fmt.Sprintf("sign unplace %d file=%s", s.LastID, file)
		p.Command(cmd)
	}

	// TODO(zchee): workaroud for "unrecovered-panic" default breakpoint.
	if id < 0 {
		id = 99
	}
	cmd := fmt.Sprintf("sign place %d name=%s line=%d file=%s", id, s.Name, line, file)
	p.Command(cmd)
	s.LastID = id
	s.LastFile = file

	return nil
}

func (s *Sign) Unplace(v *vim.Vim, id int, file string) error {
	place := fmt.Sprintf("sign unplace %d file=%s", id, file)
	if err := v.Command(place); err != nil {
		return errors.Annotate(err, pkgNvimSign)
	}

	return nil
}

func (s *Sign) UnplacePipe(p *vim.Pipeline, id int, file string) error {
	cmd := fmt.Sprintf("sign unplace %d file=%s", id, file)
	p.Command(cmd)

	return nil
}

func (s *Sign) UnplaceAll(v *vim.Vim, file string) error {
	place := fmt.Sprintf("sign unplace * file=%s", file)
	if err := v.Command(place); err != nil {
		return errors.Annotate(err, pkgNvimSign)
	}

	return nil
}

func (s *Sign) UnplaceAllPipe(p *vim.Pipeline, file string) error {
	cmd := fmt.Sprintf("sign unplace * file=%s", file)
	p.Command(cmd)

	return nil
}

func (s *Sign) UnplaceAllPcPipe(p *vim.Pipeline) error {
	cmd := fmt.Sprintf("sign unplace %d", s.LastID)
	p.Command(cmd)

	return nil
}
