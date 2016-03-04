package nvim

import (
	"strconv"
	"strings"

	"github.com/garyburd/neovim-go/vim"
)

func ParseError(v *vim.Vim, errors string) []*ErrorlistData {
	var errlist []*ErrorlistData

	el := strings.Split(errors, "\n")
	for _, es := range el {
		if e := strings.SplitN(es, ":", 3); len(e) > 1 {
			line, err := strconv.ParseInt(e[1], 10, 64)
			if err != nil {
				continue
			}
			errlist = append(errlist, &ErrorlistData{
				FileName: e[0],
				LNum:     int(line),
				Text:     e[2],
			})
		}
	}
	return errlist
}
