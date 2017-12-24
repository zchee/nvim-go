// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"nvim-go/buildctx"
	"nvim-go/nvimutil"
	"nvim-go/testutil"

	"github.com/neovim/go-client/nvim"
)

func TestCommand_Vet(t *testing.T) {
	testVetRoot := filepath.Join(testGoPath, "src", "vet")
	ctx := testutil.TestContext(context.Background())

	type fields struct {
		ctx       context.Context
		Nvim      *nvim.Nvim
		buildctxt *buildctx.Context
	}
	type args struct {
		args []string
		eval *CmdVetEval
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
		tool   string
	}{
		// method.go:17: method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error
		// method.go:21: method ReadByte() byte should have signature ReadByte() (byte, error)
		{
			name: "method.go (2 suggest)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testVetRoot, "method.go")),
				buildctxt: buildctx.NewContext(),
			},
			args: args{
				args: []string{"method.go"},
				eval: &CmdVetEval{
					Cwd:  testVetRoot,
					File: filepath.Join(testVetRoot, "method.go"),
				},
			},
			want: []*nvim.QuickfixError{{
				FileName: "method.go",
				LNum:     17,
				Col:      0,
				Text:     "method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error",
			}, {
				FileName: "method.go",
				LNum:     21,
				Col:      0,
				Text:     "method ReadByte() byte should have signature ReadByte() (byte, error)",
			}},
			tool: "go",
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
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testVetRoot, "unused.go")),
				buildctxt: buildctx.NewContext(),
			},
			args: args{
				args: []string{"."},
				eval: &CmdVetEval{
					Cwd:  testVetRoot,
					File: filepath.Join(testVetRoot, "unused.go"),
				},
			},
			want: []*nvim.QuickfixError{{
				FileName: "method.go",
				LNum:     17,
				Col:      0,
				Text:     "method Scan(x fmt.ScanState, c byte) should have signature Scan(fmt.ScanState, rune) error",
			}, {
				FileName: "method.go",
				LNum:     21,
				Col:      0,
				Text:     "method ReadByte() byte should have signature ReadByte() (byte, error)",
			}, {
				FileName: "unused.go",
				LNum:     16,
				Col:      0,
				Text:     "result of fmt.Errorf call not used",
			}, {
				FileName: "unused.go",
				LNum:     19,
				Col:      0,
				Text:     "result of errors.New call not used",
			}, {
				FileName: "unused.go",
				LNum:     22,
				Col:      0,
				Text:     "result of (error).Error call not used",
			}, {
				FileName: "unused.go",
				LNum:     25,
				Col:      0,
				Text:     "result of (bytes.Buffer).String call not used",
			}, {
				FileName: "unused.go",
				LNum:     27,
				Col:      0,
				Text:     "result of fmt.Sprint call not used",
			}, {
				FileName: "unused.go",
				LNum:     28,
				Col:      0,
				Text:     "result of fmt.Sprintf call not used",
			}},
			tool: "go",
		},
	}
	for _, tt := range tests {
		tt := tt
		tt.fields.buildctxt.Build.Tool = tt.tool
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCommand(tt.fields.ctx, tt.fields.Nvim, tt.fields.buildctxt)
			if got := c.Vet(tt.args.args, tt.args.eval); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Command.Vet(%v, %v) = %v, want %v", tt.args.args, tt.args.eval, got, tt.want)
			}
		})
	}
}
