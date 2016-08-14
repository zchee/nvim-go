package pathutil

import (
	"fmt"
	"go/build"
	"testing"
)

func TestBuildContext_PackageID(t *testing.T) {
	type fields struct {
		Tool        string
		ProjectRoot string
	}
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			fields: fields{
				Tool: "go",
			},
			args:    args{dir: astdump},
			want:    "astdump",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		switch tt.fields.Tool {
		case "go":
			build.Default.GOPATH = testGoPath
		case "gb":
			build.Default.GOPATH = fmt.Sprintf("%s:%s/vendor", projectRoot, projectRoot)
		}
		got, err := PackageID(tt.args.dir)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. BuildContext.PackagePath(%v) error = %v, wantErr %v", tt.name, tt.args.dir, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("%q. BuildContext.PackagePath(%v) = %v, want %v", tt.name, tt.args.dir, got, tt.want)
		}
	}
}
