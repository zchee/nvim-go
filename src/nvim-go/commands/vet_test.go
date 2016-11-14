// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/context"

	"github.com/neovim/go-client/nvim"
)

func TestCommands_Vet(t *testing.T) {
	type fields struct {
		Nvim     *nvim.Nvim
		Pipeline *nvim.Pipeline
		Batch    *nvim.Batch
		ctxt     *context.Context
	}
	type args struct {
		args []string
		eval *CmdVetEval
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*nvim.QuickfixError
		wantErr bool
	}{
		{
			name: "method.go(2 suggest)",
			fields: fields{
				Nvim: testVim(t, filepath.Join(testGoPath, "src/vet/method.go")),
			},
			args: args{
				args: []string{"method.go"},
				eval: &CmdVetEval{
					Cwd: filepath.Join(testGoPath, "src/vet"),
					Dir: filepath.Join(testGoPath, "src/vet"),
				},
			},
			want: []*nvim.QuickfixError{&nvim.QuickfixError{
				FileName: "method.go",
				LNum:     17,
				Col:      0,
				Text:     "method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error",
			}, &nvim.QuickfixError{
				FileName: "method.go",
				LNum:     21,
				Col:      0,
				Text:     "method ReadByte() byte should have signature ReadByte() (byte, error)",
			}},
			wantErr: false,
		},
		{
			name: "method.go + unused.go(8 suggest)",
			fields: fields{
				Nvim: testVim(t, filepath.Join(cwd, "testdata/vet/unused.go")),
			},
			args: args{
				args: []string{"."},
				eval: &CmdVetEval{
					Cwd: filepath.Join(testGoPath, "src/vet"),
					Dir: filepath.Join(testGoPath, "src/vet"),
				},
			},
			want: []*nvim.QuickfixError{&nvim.QuickfixError{
				FileName: "method.go",
				LNum:     17,
				Col:      0,
				Text:     "method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error",
			}, &nvim.QuickfixError{
				FileName: "method.go",
				LNum:     21,
				Col:      0,
				Text:     "method ReadByte() byte should have signature ReadByte() (byte, error)",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     16,
				Col:      0,
				Text:     "result of fmt.Errorf call not used",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     19,
				Col:      0,
				Text:     "result of errors.New call not used",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     22,
				Col:      0,
				Text:     "result of (error).Error call not used",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     25,
				Col:      0,
				Text:     "result of (bytes.Buffer).String call not used",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     27,
				Col:      0,
				Text:     "result of fmt.Sprint call not used",
			}, &nvim.QuickfixError{
				FileName: "unused.go",
				LNum:     28,
				Col:      0,
				Text:     "result of fmt.Sprintf call not used",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.fields.ctxt = context.NewContext()
		c := NewCommands(tt.fields.Nvim, tt.fields.ctxt)
		got, err := c.Vet(tt.args.args, tt.args.eval)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. Commands.Vet(%v, %v) error = %v, wantErr %v", tt.name, tt.args.args, tt.args.eval, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Logf("%+v", got[0])
			t.Errorf("%q. Commands.Vet(%v, %v) = %v, want %v", tt.name, tt.args.args, tt.args.eval, got, tt.want)
		}
	}
}
