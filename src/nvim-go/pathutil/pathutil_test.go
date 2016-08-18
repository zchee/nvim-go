package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neovim-go/vim"
)

var (
	cwd, _         = os.Getwd()
	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))
	testdata       = filepath.Join(projectRoot, "test", "testdata")
	testGoPath     = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
)

func testVim(t *testing.T, file string) *vim.Vim {
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

func TestChdir(t *testing.T) {
	type args struct {
		v   *vim.Vim
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantCwd string
	}{
		{
			name: "nvim-go (gb)",
			args: args{
				v:   testVim(t, projectRoot),
				dir: filepath.Join(projectRoot, "src", "nvim-go"),
			},
			wantCwd: filepath.Join(projectRoot, "src", "nvim-go"),
		},
	}
	for _, tt := range tests {
		defer func() {
			if cwd != filepath.Join(projectRoot, "src/nvim-go/pathutil") || cwd == tt.args.dir {
				t.Errorf("%q. Chdir(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.dir, cwd, tt.wantCwd)
			}
		}()
		defer Chdir(tt.args.v, tt.args.dir)()
		var ccwd interface{}
		tt.args.v.Eval("getcwd()", &ccwd)
		if ccwd.(string) != tt.wantCwd {
			t.Errorf("%q. Chdir(%v, %v) = %v, want %v", tt.name, tt.args.v, tt.args.dir, ccwd, tt.wantCwd)
		}
	}
}

func TestRel(t *testing.T) {
	type args struct {
		f   string
		cwd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{f: filepath.Join(cwd, "pathutil_test.go"), cwd: cwd},
			want: "pathutil_test.go",
		},
		{
			args: args{f: filepath.Join(cwd, "pathutil_test.go"), cwd: projectRoot},
			want: "src/nvim-go/pathutil/pathutil_test.go",
		},
	}
	for _, tt := range tests {
		if got := Rel(tt.args.f, tt.args.cwd); got != tt.want {
			t.Errorf("%q. Rel(%v, %v) = %v, want %v", tt.name, tt.args.f, tt.args.cwd, got, tt.want)
		}
	}
}

func TestExpandGoRoot(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := ExpandGoRoot(tt.args.p); got != tt.want {
			t.Errorf("%q. ExpandGoRoot(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}

func TestIsDir(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{filename: cwd},
			want: true,
		},
		{
			args: args{filename: filepath.Join(cwd, "pathutil_test.go")},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := IsDir(tt.args.filename); got != tt.want {
			t.Errorf("%q. IsDir(%v) = %v, want %v", tt.name, tt.args.filename, got, tt.want)
		}
	}
}

func TestIsExist(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{filename: cwd},
			want: true,
		},
		{
			args: args{filename: filepath.Join(cwd, "pathutil_test.go")},
			want: true,
		},
		{
			args: args{filename: filepath.Join(cwd, "not_exist.go")},
			want: false,
		},
	}
	for _, tt := range tests {
		if got := IsExist(tt.args.filename); got != tt.want {
			t.Errorf("%q. IsExist(%v) = %v, want %v", tt.name, tt.args.filename, got, tt.want)
		}
	}
}
