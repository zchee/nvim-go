// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvim

import (
	"fmt"

	vim "github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

const pkgNvimSign = "nvim.sign"

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

// Sign represents a Neovim sign.
type Sign struct {
	Name   string
	Text   string
	Texthl string
	Linehl string

	LastID   int
	LastLine int
	LastFile string
}

// NewSign define new sign and return the Sign type structure.
func NewSign(v *vim.Nvim, name, text, texthl, linehl string) (*Sign, error) {
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

// Place places the sign to any file.
func (s *Sign) Place(v *vim.Nvim, id, line int, file string, clearLastSign bool) error {
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
		return errors.Wrap(err, pkgNvimSign)
	}
	s.LastID = id
	s.LastFile = file

	return nil
}

// Unplace unplace the sign.
func (s *Sign) Unplace(v *vim.Nvim, id int, file string) error {
	place := fmt.Sprintf("sign unplace %d file=%s", id, file)
	if err := v.Command(place); err != nil {
		return errors.Wrap(err, pkgNvimSign)
	}

	return nil
}

// UnplaceAll unplace all sign on any file.
func (s *Sign) UnplaceAll(v *vim.Nvim, file string) error {
	place := fmt.Sprintf("sign unplace * file=%s", file)
	if err := v.Command(place); err != nil {
		return errors.Wrap(err, pkgNvimSign)
	}

	return nil
}
