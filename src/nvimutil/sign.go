// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"fmt"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

const (
	// BreakpointSymbol symbol of breakpoint.
	//
	// ●  BLACK CIRCLE                         (U+25CF)
	BreakpointSymbol = "\u25cf"
	// BreakpointSymbolLarge symbol of breakpoint with large.
	//
	// ⬤  BLACK LARGE CIRCLE                   (U+2B24)
	BreakpointSymbolLarge = "\u2b24"
	// TracepointSymbol symbol of tracepoint.
	//
	// ◆  BLACK DIAMOND                        (U+25C6)
	TracepointSymbol = "\u25c6"
	// TracepointSymbolMidium symbol of tracepoint with midium.
	//
	// ⬥  BLACK DIAMOND SUIT                   (U+2B25)
	TracepointSymbolMidium = "\u2b25"
	// ProgramCounterSymbol symbol of program counter.
	//
	// ◎  BULLSEYE                             (U+25CE)
	ProgramCounterSymbol = "\u25ce"
	// ProgramCounterSymbolRing symbol of program counter with ring.
	//
	// ⏣  BENZENE RING WITH CIRCLE             (U+23e3)
	ProgramCounterSymbolRing = "\u23e3"
	// ErrorSymbol symbol of error.
	//
	//  ⃠  COMBINING ENCLOSING CIRCLE BACKSLASH (U+20E0)
	ErrorSymbol = "\u20e0"
	// RestartSymbol symbol of restart.
	// ⟲  ANTICLOCKWISE GAPPED CIRCLE ARROW    (U+27F2)
	RestartSymbol = "\u27f2"
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
func NewSign(v *nvim.Nvim, name, text, texthl, linehl string) (*Sign, error) {
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
func (s *Sign) Place(v *nvim.Nvim, id, line int, file string, clearLastSign bool) error {
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
		return errors.WithStack(err)
	}
	s.LastID = id
	s.LastFile = file

	return nil
}

// Unplace unplace the sign.
func (s *Sign) Unplace(v *nvim.Nvim, id int, file string) error {
	place := fmt.Sprintf("sign unplace %d file=%s", id, file)
	if err := v.Command(place); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UnplaceAll unplace all sign on any file.
func (s *Sign) UnplaceAll(v *nvim.Nvim, file string) error {
	place := fmt.Sprintf("sign unplace * file=%s", file)
	if err := v.Command(place); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
