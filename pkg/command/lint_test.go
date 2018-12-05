// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neovim/go-client/nvim"

	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/nvimutil"
	"github.com/zchee/nvim-go/pkg/internal/testutil"
)

func TestCommand_Lint(t *testing.T) {
	testLintDir := filepath.Join("../testdata", "go", "src", "lint")
	ctx := testutil.TestContext(t, context.Background())

	type fields struct {
		ctx       context.Context
		Nvim      *nvim.Nvim
		bctxt *buildctxt.Context
	}
	type args struct {
		args []string
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*nvim.QuickfixError
		wantErr bool
	}{
		{
			name: "no suggest(go)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testLintDir, "blank-import-main.go")),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				file: filepath.Join(testLintDir, "blank-import-main.go"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "2 suggest(go)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testLintDir, "time.go")),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				args: []string{"%"},
				file: filepath.Join(testLintDir, "time.go"),
			},
			want: []*nvim.QuickfixError{
				{
					FileName: "time.go",
					LNum:     11,
					Col:      5,
					Text:     "var rpcTimeoutMsec is of type *time.Duration; don't use unit-specific suffix \"Msec\"",
				}, {
					FileName: "time.go",
					LNum:     13,
					Col:      5,
					Text:     "var timeoutSecs is of type time.Duration; don't use unit-specific suffix \"Secs\"",
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// TODO(zchee): fix lint behaiviour
			t.Skipf("TODO(zchee): fix lint behaiviour")
			t.Parallel()

			ctx := tt.fields.ctx
			c := NewCommand(ctx, tt.fields.Nvim, tt.fields.bctxt)
			c.Nvim.SetCurrentDirectory(filepath.Dir(tt.args.file))

			err := c.Lint(ctx, tt.args.args, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.Lint(%v, %v) error = %v, wantErr %v", tt.args.args, tt.args.file, err, tt.wantErr)
			}
		})
	}
}

func TestCommand_cmdLintComplete(t *testing.T) {
	testLintDir := filepath.Join("../testdata", "go", "src", "lint")
	ctx := testutil.TestContext(t, context.Background())

	type fields struct {
		ctx       context.Context
		Nvim      *nvim.Nvim
		bctxt *buildctxt.Context
	}
	type args struct {
		a   *nvim.CommandCompletionArgs
		cwd string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFilelist []string
		wantErr      bool
	}{
		{
			name: "lint dir - no args (go)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testLintDir, "make.go")),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: testLintDir,
			},
			wantFilelist: []string{"blank-import-main.go", "make.go", "time.go"},
		},
		{
			name: "lint dir - 'ma' (go)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, filepath.Join(testLintDir, "make.go")),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				a: &nvim.CommandCompletionArgs{
					ArgLead: "ma",
				},
				cwd: testLintDir,
			},
			wantFilelist: []string{"make.go"},
		},
		{
			name: "astdump (go)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, astdumpMain),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: astdump,
			},
			wantFilelist: []string{"astdump.go"},
		},
		{
			name: "gsftp (gb)",
			fields: fields{
				ctx:       ctx,
				Nvim:      nvimutil.TestNvim(t, gsftpMain),
				bctxt: buildctxt.NewContext(),
			},
			args: args{
				a:   new(nvim.CommandCompletionArgs),
				cwd: gsftp,
			},
			wantFilelist: []string{"main.go"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewCommand(tt.fields.ctx, tt.fields.Nvim, tt.fields.bctxt)
			gotFilelist, err := c.cmdLintComplete(tt.args.a, tt.args.cwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. Commands.cmdLintComplete(%v, %v) error = %v, wantErr %v", tt.name, tt.args.a, tt.args.cwd, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFilelist, tt.wantFilelist) {
				t.Errorf("%q. Commands.cmdLintComplete(%v, %v) = %v, want %v", tt.name, tt.args.a, tt.args.cwd, gotFilelist, tt.wantFilelist)
			}
		})
	}
}
