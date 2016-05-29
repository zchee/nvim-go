package commands

import (
	"go/token"
	"nvim-go/nvim/quickfix"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/garyburd/neovim-go/vim"
)

func TestGuru(t *testing.T) {
	tests := []struct {
		// Parameters.
		v    *vim.Vim
		args []string
		eval *funcGuruEval
		// Expected results.
		wantErr bool
	}{
		{
			v:    testVim(t, gsftpMain),
			args: []string{"definition"},
			eval: &funcGuruEval{
				Cwd:      gsftp,
				File:     gsftpMain,
				Modified: 0,
			},
		},
	}
	for _, tt := range tests {
		if err := Guru(tt.v, tt.args, tt.eval); (err != nil) != tt.wantErr {
			t.Errorf("Guru(%v, %v, %v) error = %v, wantErr %v", tt.v, tt.args, tt.eval, err, tt.wantErr)
		}
	}
}

func BenchmarkGuru(b *testing.B) {
	xdgDataHome := filepath.Join(testdata, "local", "share")
	os.Setenv("XDG_DATA_HOME", xdgDataHome)
	os.Setenv("NVIM_GO_DEBUG", "")
	v := benchVim(b, gsftpMain)
	w, err := v.CurrentWindow()
	if err != nil {
		b.Errorf("%v", err)
	}
	v.SetWindowCursor(w, [2]int{106, 26})
	line, _ := v.CurrentLine()
	b.Logf("%s", string(line))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Guru(v, []string{"definition"}, &funcGuruEval{
			Cwd:      gsftp,
			File:     gsftpMain,
			Modified: 0,
		}); err != nil {
			b.Errorf("BenchmarkBuildGo: %v", err)
		}
	}
}

func TestParseResult(t *testing.T) {
	tests := []struct {
		// Parameters.
		mode string
		fset *token.FileSet
		data []byte
		cwd  string
		// Expected results.
		want    []*quickfix.ErrorlistData
		want1   string
		want2   int
		want3   int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, got1, got2, got3, err := parseResult(tt.mode, tt.fset, tt.data, tt.cwd)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseResult(%v, %v, %v, %v) error = %v, wantErr %v", tt.mode, tt.fset, tt.data, tt.cwd, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseResult(%v, %v, %v, %v) = %v, want %v", tt.mode, tt.fset, tt.data, tt.cwd, got, tt.want)
		}
		if !reflect.DeepEqual(got1, tt.want1) {
			t.Errorf("parseResult(%v, %v, %v, %v) = %v, want %v", tt.mode, tt.fset, tt.data, tt.cwd, got, tt.want1)
		}
		if !reflect.DeepEqual(got2, tt.want2) {
			t.Errorf("parseResult(%v, %v, %v, %v) = %v, want %v", tt.mode, tt.fset, tt.data, tt.cwd, got, tt.want2)
		}
		if !reflect.DeepEqual(got3, tt.want3) {
			t.Errorf("parseResult(%v, %v, %v, %v) = %v, want %v", tt.mode, tt.fset, tt.data, tt.cwd, got, tt.want3)
		}
	}
}

func TestGuruHelp(t *testing.T) {
	tests := []struct {
		// Parameters.
		v    *vim.Vim
		mode string
		// Expected results.
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if err := guruHelp(tt.v, tt.mode); (err != nil) != tt.wantErr {
			t.Errorf("guruHelp(%v, %v) error = %v, wantErr %v", tt.v, tt.mode, err, tt.wantErr)
		}
	}
}
