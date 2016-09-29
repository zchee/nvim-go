package commands

import (
	"os"
	"path/filepath"
	"testing"

	vim "github.com/neovim/go-client/nvim"
)

var (
	cwd, _ = os.Getwd()

	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))
	testdata       = filepath.Join(projectRoot, "test", "testdata")
	testGoPath     = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
	broken      = filepath.Join(testGoPath, "src", "broken")
	brokenMain  = filepath.Join(astdump, "broken.go")
	gsftp       = filepath.Join(testdata, "gb", "gsftp", "src", "cmd", "gsftp")
	gsftpRoot   = filepath.Join(testdata, "gb", "gsftp")
	gsftpMain   = filepath.Join(gsftpRoot, "src", "cmd", "gsftp", "main.go")
)

func testVim(t *testing.T, file string) *vim.Nvim {
	tmpdir := filepath.Join(os.TempDir(), "nvim-go-test")
	setXDGEnv(tmpdir)
	defer os.RemoveAll(tmpdir)

	os.Setenv("NVIM_GO_DEBUG", "")

	// -u: Use <init.vim> instead of the default
	// -n: No swap file, use memory only
	nvimArgs := []string{"-u", "NONE", "-n"}
	if file != "" {
		nvimArgs = append(nvimArgs, file)
	}
	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: nvimArgs,
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}

	go v.Serve()
	return v
}

func benchVim(b *testing.B, file string) *vim.Nvim {
	tmpdir := filepath.Join(os.TempDir(), "nvim-go-test")
	setXDGEnv(tmpdir)
	defer os.RemoveAll(tmpdir)

	os.Setenv("NVIM_GO_DEBUG", "")

	// -u: Use <init.vim> instead of the default
	// -n: No swap file, use memory only
	nvimArgs := []string{"-u", "NONE", "-n"}
	if file != "" {
		nvimArgs = append(nvimArgs, file)
	}
	v, err := vim.NewEmbedded(&vim.EmbedOptions{
		Args: nvimArgs,
		Logf: b.Logf,
	})
	if err != nil {
		b.Fatal(err)
	}

	go v.Serve()
	return v
}

func setXDGEnv(tmpdir string) {
	xdgDir := filepath.Join(tmpdir, "xdg")
	os.MkdirAll(xdgDir, 0)

	os.Setenv("XDG_RUNTIME_DIR", xdgDir)
	os.Setenv("XDG_DATA_HOME", xdgDir)
	os.Setenv("XDG_CONFIG_HOME", xdgDir)
	os.Setenv("XDG_DATA_DIRS", xdgDir)
	os.Setenv("XDG_CONFIG_DIRS", xdgDir)
	os.Setenv("XDG_CACHE_HOME", xdgDir)
	os.Setenv("XDG_LOG_HOME", xdgDir)
}
