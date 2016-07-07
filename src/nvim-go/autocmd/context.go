package autocmd

import (
	"nvim-go/context"
	"nvim-go/nvim/quickfix"

	"github.com/garyburd/neovim-go/vim/plugin"
)

type AutocmdContext struct {
	ctxt *context.Context

	qf []*quickfix.ErrorlistData

	bufWritePostChan chan error
	bufWritePreChan  chan error
}

func init() {
	autocmdContext := new(AutocmdContext)

	autocmdContext.bufWritePostChan = make(chan error, 8)
	autocmdContext.bufWritePreChan = make(chan error, 1)

	plugin.HandleAutocmd("BufWritePre",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p')]"}, autocmdContext.autocmdBufWritePre)
	plugin.HandleAutocmd("BufWritePost",
		&plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go", Eval: "[getcwd(), expand('%:p:h')]"}, autocmdContext.autocmdBufWritePost)
}

func (a *AutocmdContext) send(ch chan error, fn error) {
	ch <- fn
}
