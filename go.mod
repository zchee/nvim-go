module github.com/zchee/nvim-go

go 1.13

require (
	cloud.google.com/go v0.47.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/DataDog/opencensus-go-exporter-datadog v0.0.0-20191104151809-4edbf97b176f
	github.com/aws/aws-sdk-go v1.25.29 // indirect
	github.com/cweill/gotests v1.5.4-0.20190630173305-a871e1d1c88b
	github.com/davecgh/go-spew v1.1.1
	github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	github.com/fatih/color v1.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/hokaccha/go-prettyjson v0.0.0-20190818114111-108c894c2c0e
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/motemen/go-astmanip v0.0.0-20160104081417-d6ad31f02153
	github.com/neovim/go-client v1.0.0
	github.com/peterh/liner v1.1.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.3.0 // indirect
	github.com/reviewdog/errorformat v0.0.0-20190922193611-a885a245ae0b
	github.com/tinylib/msgp v1.1.1-0.20191014034648-490d90d0e691 // indirect
	github.com/zchee/color/v2 v2.0.3
	github.com/zchee/go-xdgbasedir v1.0.3
	go.opencensus.io v0.22.1
	go.uber.org/multierr v1.4.0
	go.uber.org/zap v1.12.0
	golang.org/x/arch v0.0.0-20170711125641-f40095975f84 // indirect
	golang.org/x/debug v0.0.0-20160621010512-fb508927b491 // indirect
	golang.org/x/exp/errors v0.0.0-20191030013958-a1ab85dbe136
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de
	golang.org/x/net v0.0.0-20191105084925-a882066a44e0 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd
	golang.org/x/tools v0.0.0-20191107010934-f79515f33823
	google.golang.org/api v0.13.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.25.0 // indirect
	gopkg.in/DataDog/dd-trace-go.v1 v1.19.0 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
	gopkg.in/yaml.v3 v3.0.0-20191106092431-e228e37189d3
)

// pin delve and dependency packages
replace (
	github.com/go-delve/delve v1.2.0 => github.com/derekparker/delve v0.12.3-0.20170419170936-92dad944d7e0
	golang.org/x/arch => golang.org/x/arch v0.0.0-20170711125641-f40095975f84
	golang.org/x/debug => golang.org/x/debug v0.0.0-20160621010512-fb508927b491
)
