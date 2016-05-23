// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package highlight

import (
	"fmt"
	"log"
	"nvim-go/nvim"
	"time"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
)

type Fade struct {
	v              *vim.Vim
	buffer         vim.Buffer
	hlGroup        string
	startLine      int
	endLine        int
	startCol       int
	endCol         int
	duration       time.Duration
	timingFunction string // WIP
}

func NewFader(v *vim.Vim, buffer vim.Buffer, hlGroup string, startLine, endLine, startCol, endCol int, duration int) *Fade {
	return &Fade{
		v:         v,
		buffer:    buffer,
		hlGroup:   hlGroup,
		startLine: startLine,
		endLine:   endLine,
		startCol:  startCol,
		endCol:    endCol,
		duration:  time.Duration(int64(duration)),
	}
}

func (f *Fade) SetHighlight() error {
	if f.startLine == f.endLine {
		if _, err := f.v.AddBufferHighlight(f.buffer, 0, f.hlGroup, f.startLine, f.startCol, f.endCol); err != nil {
			return nvim.ErrorWrap(f.v, errors.Annotate(err, "highlight.FadeOut"))
		}
		chex, _ := f.v.NameToColor(f.hlGroup)
		log.Printf("chex: %s\n", chex)
		return nil
	}

	for i := f.startLine; i < f.endLine; i++ {
		if _, err := f.v.AddBufferHighlight(f.buffer, 0, f.hlGroup, f.startLine, f.startCol, f.endCol); err != nil {
			return nvim.ErrorWrap(f.v, errors.Annotate(err, "highlight.FadeOut"))
		}
		chex, _ := f.v.NameToColor(f.hlGroup)
		log.Printf("chex: %s\n", chex)
	}
	return nil
}

func (f *Fade) FadeOut() error {
	var srcID int

	for i := 0; i < 5; i++ {
		if srcID != 0 {
			f.v.ClearBufferHighlight(f.buffer, srcID, f.startLine, -1)
		}
		srcID, _ = f.v.AddBufferHighlight(f.buffer, 0, fmt.Sprintf("%s%d", f.hlGroup, i+1), f.startLine, f.startCol, f.endCol)

		time.Sleep(time.Duration(f.duration * time.Millisecond))
	}
	f.v.ClearBufferHighlight(f.buffer, srcID, f.startLine, -1)

	return nil
}
