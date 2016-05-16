package nvim

import (
	"fmt"

	"github.com/garyburd/neovim-go/vim"
)

type Sign struct {
	Id   int
	Name string
	Line int
	File string
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
		Name: name,
	}, nil
}

// func (s *S

func (s *Sign) Place(p *vim.Pipeline, id, line int, file string, clearLast bool) error {
	if clearLast && s.Id != 0 {
		cmd := fmt.Sprintf("sign unplace %d file=%s", s.Id, file)
		p.Command(cmd)
	}

	if id <= 0 {
		id = 99
	}
	cmd := fmt.Sprintf("sign place %d name=%s line=%d file=%s", id, s.Name, line, file)
	p.Command(cmd)
	s.Id = id

	return nil
}

func (s *Sign) Unplace(p *vim.Pipeline, id int, file string) error {
	cmd := fmt.Sprintf("sign unplace %d file=%s", id, file)
	p.Command(cmd)

	return nil
}

func (s *Sign) UnplaceAll(p *vim.Pipeline, file string) error {
	cmd := fmt.Sprintf("sign unplace * file=%s", file)
	p.Command(cmd)

	return nil
}
