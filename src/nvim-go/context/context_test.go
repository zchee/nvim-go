package context

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/neovim-go/vim"
	"golang.org/x/net/context"
)

var (
	cwd, _ = os.Getwd()

	projectRoot, _ = filepath.Abs(filepath.Join(cwd, "../../.."))
	testdata       = filepath.Join(projectRoot, "test", "testdata")
	testGoPath     = filepath.Join(testdata, "go")

	astdump     = filepath.Join(testGoPath, "src", "astdump")
	astdumpMain = filepath.Join(astdump, "astdump.go")
)

func TestNewContext(t *testing.T) {
	tests := []struct {
		name string
		want *Context
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := NewContext(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NewContext() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestContext_buildContext(t *testing.T) {
	type fields struct {
		Context context.Context
		Build   Build
		Errlist map[string][]*vim.QuickfixError
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
		ctxt := &Context{
			Context: tt.fields.Context,
			Build:   tt.fields.Build,
			Errlist: tt.fields.Errlist,
		}
		if got := ctxt.buildContext(tt.args.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Context.buildContext(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}

func TestContext_SetContext(t *testing.T) {
	type fields struct {
		Context context.Context
		Build   Build
		Errlist map[string][]*vim.QuickfixError
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
		ctxt := &Context{
			Context: tt.fields.Context,
			Build:   tt.fields.Build,
			Errlist: tt.fields.Errlist,
		}
		if got := ctxt.SetContext(tt.args.p); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Context.SetContext(%v) = %v, want %v", tt.name, tt.args.p, got, tt.want)
		}
	}
}
