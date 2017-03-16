// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/context"
	"nvim-go/nvimutil"

	"github.com/neovim/go-client/nvim"
)

var testVetRoot = filepath.Join(testGoPath, "src", "vet")

func TestCommands_Vet(t *testing.T) {
	type fields struct {
		Nvim  *nvim.Nvim
		Batch *nvim.Batch
		ctxt  *context.Context
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
		tool    string
	}{
		// method.go:17: method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error
		// method.go:21: method ReadByte() byte should have signature ReadByte() (byte, error)
		{
			name: "method.go (2 suggest)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testVetRoot, "method.go")),
				ctxt: context.NewContext(),
			},
			args: args{
				args: []string{"method.go"},
				eval: &CmdVetEval{
					Cwd:  testVetRoot,
					File: filepath.Join(testVetRoot, "method.go"),
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
			tool:    "go",
		},

		// method.go:17: method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error
		// method.go:21: method ReadByte() byte should have signature ReadByte() (byte, error)
		// unused.go:16: result of fmt.Errorf call not used
		// unused.go:19: result of errors.New call not used
		// unused.go:22: result of (error).Error call not used
		// unused.go:25: result of (bytes.Buffer).String call not used
		// unused.go:27: result of fmt.Sprint call not used
		// unused.go:28: result of fmt.Sprintf call not used
		{
			name: "method.go and unused.go(8 suggest)",
			fields: fields{
				Nvim: nvimutil.TestNvim(t, filepath.Join(testVetRoot, "unused.go")),
				ctxt: context.NewContext(),
			},
			args: args{
				args: []string{"."},
				eval: &CmdVetEval{
					Cwd:  testVetRoot,
					File: filepath.Join(testVetRoot, "unused.go"),
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
			tool:    "go",
		},
	}
	for _, tt := range tests {
		tt.fields.ctxt.Build.Tool = tt.tool
		c := NewCommands(tt.fields.Nvim, tt.fields.ctxt)

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := c.Vet(tt.args.args, tt.args.eval)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. Commands.Vet(%v, %v) error = %v, wantErr %v", tt.name, tt.args.args, tt.args.eval, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. Commands.Vet(%v, %v) =", tt.name, tt.args.args, tt.args.eval)
				for _, g := range got {
					t.Errorf("%+v", g)
				}
				t.Error("want =")
				for _, w := range tt.want {
					t.Errorf("%+v", w)
				}
			}
		})
	}
}
