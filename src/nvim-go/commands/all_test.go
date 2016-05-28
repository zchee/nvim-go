package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/garyburd/neovim-go/vim"
)

var (
	home           = "/Users/zchee"
	goPath         = os.Getenv("GOPATH")
	cwd, _         = os.Getwd()
	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))

	testdata   = filepath.Join(projectRoot, "tests/testdata")
	testGoPath = filepath.Join(testdata, "gopath")

	astdump = filepath.Join(testGoPath, "src/astdump")
	broken  = filepath.Join(testGoPath, "src/broken")
	gsftp   = filepath.Join(testdata, "gsftp")
)

var testVim = func(t *testing.T, file string) *vim.Vim {
	v, err := vim.StartEmbeddedVim(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n", file},
		Env:  []string{},
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

var benchVim = func(b *testing.B) *vim.Vim {
	xdg_data_home := filepath.Join(testdata, "local", "share")
	os.Setenv("XDG_DATA_HOME", xdg_data_home)
	os.Setenv("NVIM_GO_DEBUG", "")

	v, err := vim.StartEmbeddedVim(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n"},
		Env:  []string{},
		Logf: b.Logf,
	})
	if err != nil {
		b.Fatal(err)
	}
	return v
}
