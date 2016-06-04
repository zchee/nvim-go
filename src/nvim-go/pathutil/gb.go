package pathutil

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
)

func GbProjectName(p, gbProjectDir string) string {
	pkgPath := strings.Replace(p, filepath.Join(gbProjectDir, "src")+string(filepath.Separator), "", 1)
	return strings.Split(pkgPath, string(filepath.Separator))[0]
}

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
