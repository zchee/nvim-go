// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import "testing"

func TestFindVCSRoot(t *testing.T) {
	type args struct {
		basedir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nvim-go (gb)",
			args: args{basedir: cwd},
			want: projectRoot,
		},
	}
	for _, tt := range tests {
		if got := FindVCSRoot(tt.args.basedir); got != tt.want {
			t.Errorf("%q. FindVCSRoot(%v) = %v, want %v", tt.name, tt.args.basedir, got, tt.want)
		}
	}
}
