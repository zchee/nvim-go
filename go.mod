module github.com/zchee/nvim-go

go 1.13

require (
	cloud.google.com/go v0.54.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8
	github.com/DataDog/datadog-go v3.4.0+incompatible // indirect
	github.com/DataDog/opencensus-go-exporter-datadog v0.0.0-20191104151809-4edbf97b176f
	github.com/cweill/gotests v1.5.4-0.20190630173305-a871e1d1c88b
	github.com/davecgh/go-spew v1.1.1
	github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/motemen/go-astmanip v0.0.0-20160104081417-d6ad31f02153
	github.com/neovim/go-client v1.0.0
	github.com/peterh/liner v1.2.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.4.0 // indirect
	github.com/reviewdog/errorformat v0.0.0-20190922193611-a885a245ae0b
	github.com/tinylib/msgp v1.1.1 // indirect
	github.com/zchee/color/v2 v2.0.3
	github.com/zchee/go-xdgbasedir v1.0.3
	go.opencensus.io v0.22.3
	go.uber.org/multierr v1.4.0
	go.uber.org/zap v1.12.0
	golang.org/x/arch v0.0.0-20170711125641-f40095975f84 // indirect
	golang.org/x/debug v0.0.0-20160621010512-fb508927b491 // indirect
	golang.org/x/exp/errors v0.0.0-20191030013958-a1ab85dbe136
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527
	golang.org/x/tools v0.0.0-20200305224536-de023d59a5d1
	gopkg.in/DataDog/dd-trace-go.v1 v1.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20191106092431-e228e37189d3
)

// pin delve and dependency packages
replace (
	github.com/go-delve/delve v1.2.0 => github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	golang.org/x/arch => golang.org/x/arch v0.0.0-20170711125641-f40095975f84
	golang.org/x/debug => golang.org/x/debug v0.0.0-20160621010512-fb508927b491
)
