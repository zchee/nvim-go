module github.com/zchee/nvim-go

go 1.15

require (
	cloud.google.com/go v0.64.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.3
	github.com/DataDog/opencensus-go-exporter-datadog v0.0.0-20200406135749-5c268882acf0
	github.com/cweill/gotests v1.5.4-0.20200413045357-2435ae532b97
	github.com/davecgh/go-spew v1.1.1
	github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/go-cmp v0.5.1
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/motemen/go-astmanip v0.0.0-20160104081417-d6ad31f02153
	github.com/neovim/go-client v1.1.2
	github.com/peterh/liner v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.5.0 // indirect
	github.com/reviewdog/errorformat v0.0.0-20200813150006-94458edd948a
	github.com/zchee/color/v2 v2.0.3
	github.com/zchee/go-xdgbasedir v1.0.3
	go.opencensus.io v0.22.4
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.15.0
	golang.org/x/arch v0.0.0-20170711125641-f40095975f84 // indirect
	golang.org/x/debug v0.0.0-20160621010512-fb508927b491 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/sync v0.0.0-20201008141435-b3e1573b7520
	golang.org/x/sys v0.0.0-20200819141100-7c7a22168250
	golang.org/x/tools v0.0.0-20201017001424-6003fad69a88
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

// pin delve and dependency packages
replace (
	github.com/go-delve/delve v1.2.0 => github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	golang.org/x/arch => golang.org/x/arch v0.0.0-20170711125641-f40095975f84
	golang.org/x/debug => golang.org/x/debug v0.0.0-20160621010512-fb508927b491
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200819141100-7c7a22168250
)
