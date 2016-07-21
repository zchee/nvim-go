package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/neovim-go/vim"
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
		return nil
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

func Expand(p string) string {
	switch {
	case strings.Index(p, "$GOROOT") != 1:
		return strings.Replace(p, "$GOROOT", runtime.GOROOT(), 1)
	}

	return p // Not hit
}

func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
