package pathutil

import (
	"path/filepath"
	"strings"
	"sync"

	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
)

var pkgPathutil = "pathutil"

// Chdir changes the vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func Chdir(v *vim.Vim, dir string) func() {
	var (
		m   sync.Mutex
		cwd interface{}
	)
	m.Lock()
	if err := v.Eval("getcwd()", &cwd); err != nil {
		nvim.Echoerr(v, "GoTerminal: %v", err)
	}
	v.ChangeDirectory(dir)
	return func() {
		v.ChangeDirectory(cwd.(string))
		m.Unlock()
	}
}

func RelPath(f, cwd string) string {
	if filepath.HasPrefix(f, cwd) {
		return strings.Replace(f, cwd+string(filepath.Separator), "", 1)
	}
	rel, _ := filepath.Rel(cwd, f)
	return rel
}
