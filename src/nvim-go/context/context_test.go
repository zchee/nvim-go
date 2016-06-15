// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"reflect"
	"testing"
)

func TestBuildContext(t *testing.T) {
	tests := []struct {
		// Receiver fields.
		rTool    string
		rContext build.Context
		// Parameters.
		p string
		// Expected results.
		want  string
		want1 string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		ctxt := &Build{
			Tool:         tt.rTool,
			BuildContext: tt.rContext,
		}
		got, got1 := ctxt.buildContext(tt.p)
		if got != tt.want {
			t.Errorf("Build.buildContext(%v) got = %v, want %v", tt.p, got, tt.want)
		}
		if got1 != tt.want1 {
			t.Errorf("Build.buildContext(%v) got1 = %v, want %v", tt.p, got1, tt.want1)
		}
	}
}

func TestSetContext(t *testing.T) {
	tests := []struct {
		// Receiver fields.
		rTool    string
		rContext build.Context
		// Parameters.
		p string
		// Expected results.
		want func()
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		c := &Build{
			Tool:         tt.rTool,
			BuildContext: tt.rContext,
		}
		if got := c.SetContext(tt.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Build.SetContext(%v) = %v, want %v", tt.p, got, tt.want)
		}
	}
}
