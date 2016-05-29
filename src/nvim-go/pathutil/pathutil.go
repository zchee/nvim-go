package pathutil

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"nvim-go/context"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/juju/errors"
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

func PackagePath(p string) string {
	goPath := os.Getenv("GOPATH")
	p = context.FindVcsRoot(p)

	return strings.Replace(p, goPath+string(filepath.Separator), "", 1)
}

func GbProjectName(p, projectDir string) string {
	pkgPath := strings.Replace(p, filepath.Join(projectDir, "src")+string(filepath.Separator), "", 1)
	return strings.Split(pkgPath, string(filepath.Separator))[0]
}

var (
	vcsDirs     = []string{".git", ".svn", ".hg"}
	vcsDirFound bool
)

func GbProject(p string, glob bool) ([]string, error) {
	var projects []string

	if filepath.Base(p) == "vendor" {
		filepath.Walk(p,
			func(path string, fileInfo os.FileInfo, err error) error {
				if err != nil || fileInfo == nil || !fileInfo.IsDir() {
					return nil
				}
				gbProject, err := build.ImportDir(path, build.ImportMode(0))
				if err != nil {
					return nil
				}
				if gbProject.Name != "" {
					projectPath := strings.Replace(gbProject.Dir, filepath.Join(p, "src")+string(filepath.Separator), "", 1)
					if glob {
						projects = append(projects, projectPath+string(filepath.Separator)+"...")
					} else {
						projects = append(projects, projectPath)
					}
				}
				return nil
			})
	} else {
		prjs, err := ioutil.ReadDir(filepath.Join(p, "src"))
		if err != nil {
			return nil, errors.Annotate(err, pkgPathutil)
		}
		for _, prj := range prjs {
			if glob {
				projects = append(projects, prj.Name()+string(filepath.Separator)+"...")
			} else {
				projects = append(projects, prj.Name())
			}
		}
	}

	return projects, nil
}
