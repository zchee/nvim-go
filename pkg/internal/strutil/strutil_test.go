// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strutil

import (
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "toCamelCase",
			s:    "test",
			want: "Test",
		},
		{
			name: "snake_case",
			s:    "test_case",
			want: "TestCase",
		},
		{
			name: "HasSpace",
			s:    " test  case ",
			want: "TestCase",
		},
		{
			name: "Empty",
			s:    "",
			want: "",
		},
		{
			name: "HasManyWords",
			s:    "many_many_words",
			want: "ManyManyWords",
		},
		{
			name: "HasAnyKind",
			s:    "AnyKind of_string",
			want: "AnyKindOfString",
		},
		{
			name: "OddFix",
			s:    "odd-fix",
			want: "OddFix",
		},
		{
			name: "HasNumber",
			s:    "numbers2And55with000",
			want: "Numbers2And55With000",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToCamelCase(tt.s); got != tt.want {
				t.Errorf("ToCamelCase(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToLowerCamelCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "hasHyphen",
			s:    "foo-bar",
			want: "fooBar",
		},
		{
			name: "UpperCamelCase",
			s:    "TestCase",
			want: "testCase",
		},
		{
			name: "Empty",
			s:    "",
			want: "",
		},
		{
			name: "HasAnyKind",
			s:    "AnyKind of_string",
			want: "anyKindOfString",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToLowerCamelCase(tt.s); got != tt.want {
				t.Errorf("ToLowerCamelCase(%v) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "lowerCamelCase",
			s:    "testCase",
			want: "test_case",
		},
		{
			name: "UpperCamelCase",
			s:    "TestCase",
			want: "test_case",
		},
		{
			name: "UpperCamelCaseWithSpace",
			s:    "Test Case",
			want: "test_case",
		},
		{
			name: "UpperCamelCasehasSpaces",
			s:    " Test Case",
			want: "test_case",
		},
		{
			name: "no-op",
			s:    "test",
			want: "test",
		},
		{
			name: "no-op2",
			s:    "test_case",
			want: "test_case",
		},
		{
			name: "Empty",
			s:    "",
			want: "",
		},
		{
			name: "HasManyWords",
			s:    "ManyManyWords",
			want: "many_many_words",
		},
		{
			name: "manyManyWords",
			s:    "manyManyWords",
			want: "many_many_words",
		},
		{
			name: "HasAnyKind",
			s:    "AnyKind of_string",
			want: "any_kind_of_string",
		},
		// {
		// 	name: "HasNumber",
		// 	s:    "numbers2and55with000",
		// 	want: "numbers_2_and_55_with_000",
		// },
		{
			name: "JSONData",
			s:    "JSONData",
			want: "json_data",
		},
		{
			name: "userID",
			s:    "userID",
			want: "user_id",
		},
		{
			name: "AAAbbb",
			s:    "AAAbbb",
			want: "aa_abbb",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToSnakeCase(tt.s); got != tt.want {
				t.Errorf("ToSnakeCase(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "lowerCamelCase",
			s:    "testCase",
			want: "TEST_CASE",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToScreamingSnakeCase(tt.s); got != tt.want {
				t.Errorf("ToScreamingSnakeCase(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "lowerCamelCase",
			s:    "testCase",
			want: "test-case",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToKebab(tt.s); got != tt.want {
				t.Errorf("ToKebab(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToScreamingKebab(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "lowerCamelCase",
			s:    "testCase",
			want: "TEST-CASE",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToScreamingKebab(tt.s); got != tt.want {
				t.Errorf("ToScreamingKebab(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestToDelimited(t *testing.T) {
	type args struct {
		s   string
		del uint8
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "lowerCamelCase",
			args: args{s: "testCase", del: '@'},
			want: "test@case",
		},
		{
			name: "UpperCamelCase",
			args: args{s: "TestCase", del: '@'},
			want: "test@case",
		},
		{
			name: "UpperCamelCaseWithSpace",
			args: args{s: "Test Case", del: '@'},
			want: "test@case",
		},
		{
			name: "UpperCamelCasehasSpaces",
			args: args{s: " Test Case", del: '@'},
			want: "test@case",
		},
		{
			name: "no-op",
			args: args{s: "test", del: '@'},
			want: "test",
		},
		{
			name: "test_case",
			args: args{s: "test_case", del: '@'},
			want: "test@case",
		},
		{
			name: "Empty",
			args: args{s: "", del: '@'},
			want: "",
		},
		{
			name: "HasManyWords",
			args: args{s: "ManyManyWords", del: '@'},
			want: "many@many@words",
		},
		{
			name: "manyManyWords",
			args: args{s: "manyManyWords", del: '@'},
			want: "many@many@words",
		},
		{
			name: "HasAnyKind",
			args: args{s: "AnyKind of_string", del: '@'},
			want: "any@kind@of@string",
		},
		// {
		// 	name: "HasNumber",
		// 	args: args{s: "numbers2and55with000", del: '@'},
		// 	want: "numbers@2@and@55@with@000",
		// },
		{
			name: "JSONData",
			args: args{s: "JSONData", del: '@'},
			want: "json@data",
		},
		{
			name: "userID",
			args: args{s: "userID", del: '@'},
			want: "user@id",
		},
		{
			name: "AAAbbb",
			args: args{s: "AAAbbb", del: '@'},
			want: "aa@abbb",
		},
		{
			name: "test-case",
			args: args{s: "test-case", del: '@'},
			want: "test@case",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToDelimited(tt.args.s, tt.args.del); got != tt.want {
				t.Errorf("ToDelimited(%v, %v) = %v, want %v", tt.args.s, tt.args.del, got, tt.want)
			}
		})
	}
}

func TestToScreamingDelimited(t *testing.T) {
	type args struct {
		s         string
		del       uint8
		screaming bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "lowerCamelCase",
			args: args{s: "testCase", del: '.', screaming: true},
			want: "TEST.CASE",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToScreamingDelimited(tt.args.s, tt.args.del, tt.args.screaming); got != tt.want {
				t.Errorf("ToScreamingDelimited(%q, %v, %v) = %v, want %v", tt.args.s, tt.args.del, tt.args.screaming, got, tt.want)
			}
		})
	}
}
