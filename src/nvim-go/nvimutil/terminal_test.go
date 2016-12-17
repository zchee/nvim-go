// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"testing"

	"github.com/neovim/go-client/nvim"
)

func TestTerminal_getWindowSize(t *testing.T) {
	v := TestNvim(t)
	type fields struct {
		v *nvim.Nvim
	}
	type args struct {
		cfg int64
		fn  func(nvim.Window) (int, error)
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
				v: v,
			},
			args: args{
				cfg: 120,
				fn:  v.WindowWidth,
			},
			want: 120,
		},
		{
			name: "Use (WindowWidth / 3 )",
			fields: fields{
				v: v,
			},
			args: args{
				cfg: 0,
				fn:  v.WindowWidth,
			},
			// default embedded Nvim Window width is 80.
			want: 26,
		},
		{
			name: "Use (WindowHeight / 3 )",
			fields: fields{
				v: v,
			},
			args: args{
				cfg: 0,
				fn:  v.WindowHeight,
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
				v: tt.fields.v,
			}
			if got := term.getWindowSize(tt.args.cfg, tt.args.fn); got != tt.want {
				t.Errorf("Terminal.getWindowSize(%v, %v) = %v, want %v", tt.args.cfg, tt.args.fn, got, tt.want)
			}
		})
	}
}
