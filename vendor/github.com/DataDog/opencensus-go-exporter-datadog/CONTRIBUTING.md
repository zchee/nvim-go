### Contributing

Pull requests for bug fixes are welcome, but before submitting new features or changes to current functionalities [open an issue](https://github.com/DataDog/opencensus-go-exporter-datadog/issues/new)
and discuss your ideas or propose the changes you wish to make. After a resolution is reached a PR can be submitted for review.

Commit messages should be prefixed with "trace" if they are related to APM or with "stats" if they are related to metrics. For example:
```
trace: add support for float64

Adds support for `float64`, added in census-instrumentation/opencensus-go#1033

Closes #37
```
Please apply the same logic for Pull Requests, start with "trace" or "stats", followed by a colon and a description of the change, similar to 
the official [Go language](https://github.com/golang/go/pulls).
