package commands

import (
	"go/token"
	"reflect"
	"testing"

	"nvim-go/internal/guru"
	"nvim-go/internal/guru/serial"
	"nvim-go/nvim/quickfix"

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
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if err := Guru(tt.v, tt.args, tt.eval); (err != nil) != tt.wantErr {
			t.Errorf("Guru(%v, %v, %v) error = %v, wantErr %v", tt.v, tt.args, tt.eval, err, tt.wantErr)
		}
	}
}

func BenchmarkGuruCallees(b *testing.B) {
	v := benchVim(b, gsftpMain)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Guru(v, []string{"callees"}, &funcGuruEval{
			Cwd:      gsftp,
			File:     gsftpMain,
			Modified: 0,
			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
		}); err != nil {
			b.Errorf(":BenchmarkGuruCallees %v", err)
		}
	}
}

func BenchmarkGuruCallers(b *testing.B) {
	v := benchVim(b, gsftpMain)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Guru(v, []string{"callers"}, &funcGuruEval{
			Cwd:      gsftp,
			File:     gsftpMain,
			Modified: 0,
			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
		}); err != nil {
			b.Errorf("BenchmarkGuruCallers: %v", err)
		}
	}
}

func TestDefinition(t *testing.T) {
	tests := []struct {
		// Parameters.
		q *guru.Query
		// Expected results.
		want    *serial.Definition
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := definition(tt.q)
		if (err != nil) != tt.wantErr {
			t.Errorf("definition(%v) error = %v, wantErr %v", tt.q, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("definition(%v) = %v, want %v", tt.q, got, tt.want)
		}
	}
}

func BenchmarkGuruDefinition(b *testing.B) {
	v := benchVim(b, gsftpMain)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Guru(v, []string{"definition"}, &funcGuruEval{
			Cwd:      gsftp,
			File:     gsftpMain,
			Modified: 0,
			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
		}); err != nil {
			b.Errorf("BenchmarkGuruDefinition: %v", err)
		}
	}
}

func TestFallbackChan(t *testing.T) {
	tests := []struct {
		// Parameters.
		obj *serial.Definition
		err error
		// Expected results.
		want fallback
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := fallbackChan(tt.obj, tt.err); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("fallbackChan(%v, %v) = %v, want %v", tt.obj, tt.err, got, tt.want)
		}
	}
}

func TestDefinitionFallback(t *testing.T) {
	tests := []struct {
		// Parameters.
		q *guru.Query
		c chan fallback
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		definitionFallback(tt.q, tt.c)
	}
}

func BenchmarkGuruDefinitionFallback(b *testing.B) {
	v := benchVim(b, gsftpMain)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Guru(v, []string{"definition"}, &funcGuruEval{
			Cwd:      gsftp,
			File:     gsftpMain,
			Modified: 0,
			Offset:   2132, // defer conn.|C|lose()
		}); err != nil {
			b.Errorf("BenchmarkGuruDefinitionFallback: %v", err)
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
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := parseResult(tt.mode, tt.fset, tt.data, tt.cwd)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseResult(%v, %v, %v, %v) error = %v, wantErr %v", tt.mode, tt.fset, tt.data, tt.cwd, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseResult(%v, %v, %v, %v) = %v, want %v", tt.mode, tt.fset, tt.data, tt.cwd, got, tt.want)
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
