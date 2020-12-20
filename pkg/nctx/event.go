// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nctx

import (
	"errors"

	"github.com/neovim/go-client/nvim"
)

const (
	EventBufLines       = "nvim_buf_lines_event"
	EventBufChangedtick = "nvim_buf_changedtick_event"
	EventBufAttach      = "nvim_buf_attach_event"
	EventBufDetach      = "nvim_buf_detach_event"
)

func RegisterEvent(n *nvim.Nvim, event string, fn func(...interface{})) {
	n.RegisterHandler(event, fn)
}

func RegisterBufLinesEvent(n *nvim.Nvim, fn func(...interface{})) {
	RegisterEvent(n, EventBufLines, fn)
}

func RegisterBufChangedtickEvent(n *nvim.Nvim, fn func(...interface{})) {
	RegisterEvent(n, EventBufChangedtick, fn)
}

func reAttachFunc(n *nvim.Nvim) (nvim.Buffer, error) {
	buf, err := n.CurrentBuffer()
	if err != nil {
		return 0, errors.New("failed to gets current buffer")
	}

	if _, err := n.AttachBuffer(buf, false, make(map[string]interface{})); err != nil {
		return 0, errors.New("failed to attach buffer")
	}

	return buf, nil
}

func RegisterBufAttachEvent(n *nvim.Nvim, fn func(...interface{})) nvim.Buffer {
	buf, err := reAttachFunc(n)
	if err != nil {
		n.WritelnErr("failed to gets current buffer")
	}

	fn()

	return buf
}

func RegisterBufDetachEvent(n *nvim.Nvim, fn func(...interface{})) {
	detachFn := func(...interface{}) {
		_, err := reAttachFunc(n)
		if err != nil {
			n.WritelnErr("failed to gets current buffer")
		}

		fn()
	}

	RegisterEvent(n, EventBufDetach, detachFn)
}
