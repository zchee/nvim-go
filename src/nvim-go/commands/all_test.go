package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neovim-go/vim"
)

var (
	goPath         = os.Getenv("GOPATH")
	cwd, _         = os.Getwd()
	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))

	testdata   = filepath.Join(projectRoot, "tests/testdata")
	testGoPath = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src/astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")

	broken     = filepath.Join(testGoPath, "src/broken")
	brokenMain = filepath.Join(astdump, "broken.go")

	gsftp     = filepath.Join(testdata, "gb", "gsftp", "src", "cmd", "gsftp")
	gsftpRoot = filepath.Join(testdata, "gb", "gsftp")
	gsftpMain = filepath.Join(gsftpRoot, "src", "cmd", "gsftp", "main.go")
)

var testVim = func(t *testing.T, file string) *vim.Vim {
	os.Setenv("NVIM_GO_DEBUG", "")

	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n", file},
		Env:  []string{},
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

var benchVim = func(b *testing.B, file string) *vim.Vim {
	xdgDataHome := filepath.Join(testdata, "local", "share")
	os.Setenv("XDG_DATA_HOME", xdgDataHome)
	os.Setenv("NVIM_GO_DEBUG", "")

	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n", file},
		Env:  []string{},
		Logf: b.Logf,
	})
	if err != nil {
		b.Fatal(err)
	}
	return v
}
