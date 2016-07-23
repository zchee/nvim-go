// Copyright 2016 Koichi Shiraishi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	cwd, _ = os.Getwd()

	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))
	testdata       = filepath.Join(projectRoot, "test", "testdata")
	testGoPath     = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
)

func TestBuildContext_buildContext(t *testing.T) {
	type fields struct {
		Tool         string
		GbProjectDir string
	}
	type args struct {
		p string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   build.Context
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		ctxt := &BuildContext{
			Tool:        tt.fields.Tool,
			ProjectRoot: tt.fields.GbProjectDir,
		}
		if got := ctxt.buildContext(tt.args.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. BuildContext.buildContext(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}

func TestBuildContext_SetContext(t *testing.T) {
	type fields struct {
		Tool         string
		GbProjectDir string
	}
	type args struct {
		p string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   func()
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		ctxt := &BuildContext{
			Tool:        tt.fields.Tool,
			ProjectRoot: tt.fields.GbProjectDir,
		}
		if got := ctxt.SetContext(tt.args.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. BuildContext.SetContext(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}
