package autocmd

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"

	"nvim-go/commands"
)

func init() {
	plugin.HandleAutocmd("BufWritePost", &plugin.AutocmdOptions{Pattern: "*.go", Eval: "*"}, autocmdBufWritePost)
}

type bufwritepostFileInfo struct {
	Cwd string `eval:"getcwd()"`
}

type bufwritepostEnv struct {
	BuildAutoSave int64 `eval:"g:go#build#autosave"`
}

func autocmdBufWritePost(v *vim.Vim, eval *struct {
	FileInfo bufwritepostFileInfo
	Env      bufwritepostEnv
}) error {
	if eval.Env.BuildAutoSave != int64(0) {
		go commands.Build(v, eval.FileInfo.Cwd)
	}
	return nil
}
