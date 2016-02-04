package buffer

import (
	"encoding/binary"

	"github.com/garyburd/neovim-go/vim"
)

func ByteOffset(v *vim.Vim) (int, error) {
	b, err := v.CurrentBuffer()
	if err != nil {
		return 0, err
	}
	w, err := v.CurrentWindow()
	if err != nil {
		return 0, err
	}

	cursor, err := v.WindowCursor(w)
	if err != nil {
		return 0, err
	}
	byteBuf, err := v.BufferLineSlice(b, 0, -1, true, true)
	if err != nil {
		return 0, err
	}

	offset := 0
	cursorline := 1
	for _, bytes := range byteBuf {
		if cursor[0] == 1 {
			offset = 1
			break
		} else if cursorline == cursor[0] {
			offset++
			break
		}
		offset += (binary.Size(bytes) + 1)
		cursorline++
	}

	return (offset + (cursor[1] - 1)), nil
}
