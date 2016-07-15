package commands

import (
	"go/token"
	"reflect"
	"testing"

	"nvim-go/context"
	"nvim-go/internal/guru"
	"nvim-go/internal/guru/serial"

	"github.com/neovim-go/vim"
)

func TestCommands_Guru(t *testing.T) {
	type fields struct {
		Vim  *vim.Vim
		p    *vim.Pipeline
		ctxt *context.Context
	}
	type args struct {
		args []string
		eval *funcGuruEval
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		c := &Commands{
			v:    tt.fields.Vim,
			p:    tt.fields.p,
			ctxt: tt.fields.ctxt,
		}
		if err := c.Guru(tt.args.args, tt.args.eval); (err != nil) != tt.wantErr {
			t.Errorf("%q. Commands.Guru(%v, %v) error = %v, wantErr %v", tt.name, tt.args.args, tt.args.eval, err, tt.wantErr)
		}
	}
}

func Test_definition(t *testing.T) {
	type args struct {
		q *guru.Query
	}
	tests := []struct {
		name    string
		args    args
		want    *serial.Definition
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := definition(tt.args.q)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. definition(%v) error = %v, wantErr %v", tt.name, tt.args.q, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. definition(%v) = %v, want %v", tt.name, tt.args.q, got, tt.want)
		}
	}
}

func Test_fallbackChan(t *testing.T) {
	type args struct {
		obj *serial.Definition
		err error
	}
	tests := []struct {
		name string
		args args
		want fallback
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := fallbackChan(tt.args.obj, tt.args.err); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. fallbackChan(%v, %v) = %v, want %v", tt.name, tt.args.obj, tt.args.err, got, tt.want)
		}
	}
}

func Test_definitionFallback(t *testing.T) {
	type args struct {
		q *guru.Query
		c chan fallback
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		definitionFallback(tt.args.q, tt.args.c)
	}
}

func Test_parseResult(t *testing.T) {
	type args struct {
		mode string
		fset *token.FileSet
		data []byte
		cwd  string
	}
	tests := []struct {
		name    string
		args    args
		want    []*vim.QuickfixError
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		got, err := parseResult(tt.args.mode, tt.args.fset, tt.args.data, tt.args.cwd)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. parseResult(%v, %v, %v, %v) error = %v, wantErr %v", tt.name, tt.args.mode, tt.args.fset, tt.args.data, tt.args.cwd, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. parseResult(%v, %v, %v, %v) = %v, want %v", tt.name, tt.args.mode, tt.args.fset, tt.args.data, tt.args.cwd, got, tt.want)
		}
	}
}

func Test_guruHelp(t *testing.T) {
	type args struct {
		v    *vim.Vim
		mode string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if err := guruHelp(tt.args.v, tt.args.mode); (err != nil) != tt.wantErr {
			t.Errorf("%q. guruHelp(%v, %v) error = %v, wantErr %v", tt.name, tt.args.v, tt.args.mode, err, tt.wantErr)
		}
	}
}

// func BenchmarkGuruDefinition(b *testing.B) {
// 	v := benchVim(b, gsftpMain)
// 	c := NewCommands(v)
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		if err := c.Guru([]string{"definition"}, &funcGuruEval{
// 			Cwd:      gsftp,
// 			File:     gsftpMain,
// 			Modified: 0,
// 			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
// 		}); err != nil {
// 			b.Errorf("BenchmarkGuruDefinition: %v", err)
// 		}
// 	}
// }
//
// func BenchmarkGuruDefinitionFallback(b *testing.B) {
// 	v := benchVim(b, gsftpMain)
// 	c := NewCommands(v)
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		if err := c.Guru([]string{"definition"}, &funcGuruEval{
// 			Cwd:      gsftp,
// 			File:     gsftpMain,
// 			Modified: 0,
// 			Offset:   2132, // defer conn.|C|lose()
// 		}); err != nil {
// 			b.Errorf("BenchmarkGuruDefinitionFallback: %v", err)
// 		}
// 	}
// }
//
// func BenchmarkGuruCallees(b *testing.B) {
// 	v := benchVim(b, gsftpMain)
// 	c := NewCommands(v)
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		if err := c.Guru([]string{"callees"}, &funcGuruEval{
// 			Cwd:      gsftp,
// 			File:     gsftpMain,
// 			Modified: 0,
// 			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
// 		}); err != nil {
// 			b.Errorf(":BenchmarkGuruCallees %v", err)
// 		}
// 	}
// }
//
// func BenchmarkGuruCallers(b *testing.B) {
// 	v := benchVim(b, gsftpMain)
// 	c := NewCommands(v)
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		if err := c.Guru([]string{"callers"}, &funcGuruEval{
// 			Cwd:      gsftp,
// 			File:     gsftpMain,
// 			Modified: 0,
// 			Offset:   2027, // client, err := sftp.|N|ewClient(conn)
// 		}); err != nil {
// 			b.Errorf("BenchmarkGuruCallers: %v", err)
// 		}
// 	}
// }
