package autocmd

import (
	"fmt"
	"log"
	"net/http"

	"nvim-go/config"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleAutocmd("BufEnter", &plugin.AutocmdOptions{Pattern: "*.go", Group: "nvim-go"}, autocmdBufEnter)
}

func autocmdBufEnter(v *vim.Vim, cwd string) error {
	if config.DebugPprof {
		fmt.Printf("Start pprof debug\n")
		go func() {
			log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
		}()
	}

	return nil
}
