// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"testing"

	"github.com/neovim/go-client/nvim"
)

func TestTerminal_getSplitWindowSize(t *testing.T) {
	n := TestNvim(t)
	type fields struct {
		Nvim *nvim.Nvim
	}
	type args struct {
		cfg int64
		f   func(nvim.Window) (int, error)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Use config",
			fields: fields{
				Nvim: n,
			},
			args: args{
				cfg: 120,
				f:   n.WindowWidth,
			},
			want: 120,
		},
		{
			name: "Use (WindowWidth / 3 )",
			fields: fields{
				Nvim: n,
			},
			args: args{
				cfg: 0,
				f:   n.WindowWidth,
			},
			// default embedded Nvim Window width is 80.
			want: 26,
		},
		{
			name: "Use (WindowHeight / 3 )",
			fields: fields{
				Nvim: n,
			},
			args: args{
				cfg: 0,
				f:   n.WindowHeight,
			},
			// default embedded Nvim Window height is 22.
			want: 7,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			term := &Terminal{
				Nvim: tt.fields.Nvim,
			}
			if got := term.getSplitWindowSize(tt.args.cfg, tt.args.f); got != tt.want {
				t.Errorf("getSplitWindowSize(%v) = %v, want %v", tt.args.cfg, got, tt.want)
			}
		})
	}
}
