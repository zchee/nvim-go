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

	testdata   = filepath.Join(projectRoot, "test", "testdata")
	testGoPath = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")

	broken     = filepath.Join(testGoPath, "src", "broken")
	brokenMain = filepath.Join(astdump, "broken.go")

	gsftp     = filepath.Join(testdata, "gb", "gsftp", "src", "cmd", "gsftp")
	gsftpRoot = filepath.Join(testdata, "gb", "gsftp")
	gsftpMain = filepath.Join(gsftpRoot, "src", "cmd", "gsftp", "main.go")
)

func testVim(t *testing.T, file string) *vim.Vim {
	xdgDataHome := filepath.Join(testdata, "local", "share")
	os.Setenv("XDG_DATA_HOME", xdgDataHome)
	os.Setenv("NVIM_GO_DEBUG", "")

	args := []string{"-u", "NONE", "-n"}
	if file != "" {
		args = append(args, file)
	}
	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: args,
		Env:  []string{},
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}

	go v.Serve()
	return v
}

func benchVim(b *testing.B, file string) *vim.Vim {
	xdgDataHome := filepath.Join(testdata, "local", "share")
	os.Setenv("XDG_DATA_HOME", xdgDataHome)
	os.Setenv("NVIM_GO_DEBUG", "")

	args := []string{"-u", "NONE", "-n"}
	if file != "" {
		args = append(args, file)
	}
	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: args,
		Env:  []string{},
		Logf: b.Logf,
	})
	if err != nil {
		b.Fatal(err)
	}

	go v.Serve()
	return v
}
