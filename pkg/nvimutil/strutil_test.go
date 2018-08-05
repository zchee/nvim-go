// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"reflect"
	"testing"
)

func TestToPascalCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "toPascalCase",
			args: args{s: "toPascalCase"},
			want: "ToPascalCase",
		},
		{
			name: "ToPascalCas",
			args: args{s: "ToPascalCase"},
			want: "ToPascalCase",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToPascalCase(tt.args.s); got != tt.want {
				t.Errorf("%q. ToPascalCase(%v) = %v, want %v", tt.name, tt.args.s, got, tt.want)
			}
		})
	}
}

func TestToMixedCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "toMixedCase",
			args: args{s: "toMixedCase"},
			want: "toMixedCase",
		},
		{
			name: "ToMixedCase",
			args: args{s: "ToMixedCase"},
			want: "toMixedCase",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToMixedCase(tt.args.s); got != tt.want {
				t.Errorf("%q. ToMixedCase(%v) = %v, want %v", tt.name, tt.args.s, got, tt.want)
			}
		})
	}
}

func TestStrToByteSlice(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ToByteSlice",
			args: args{s: "ToByteSlice"},
			want: []byte("ToByteSlice"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := StrToByteSlice(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. StringToByteslice(%v) = %v, want %v", tt.name, tt.args.s, got, tt.want)
			}
		})
	}
}
