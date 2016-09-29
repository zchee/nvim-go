package nvim

import vim "github.com/neovim/go-client/nvim"

// WindowContext represents a Neovim window context.
type WindowContext struct {
	vim.Window
}
