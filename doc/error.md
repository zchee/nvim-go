# Go typical errors

## initialization loop
```sh
# github.com/zchee/appleopensource/cmd/gaos
cmd/gaos/versions.go:14: initialization loop:
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:14 cmdVersions refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:19 runVersions refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:16 versionsPkg refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:14 cmdVersions
```

## cgo side error (go-clang)
```sh
# github.com/zchee/clang-server/vendor/github.com/go-clang/v3.9/clang
cgo-gcc-prolog:244:6: warning: 'clang_getDiagnosticCategoryName' is deprecated [-Wdeprecated-declarations]
../vendor/github.com/go-clang/v3.9/clang/clang-c/Index.h:952:10: note: 'clang_getDiagnosticCategoryName' has been explicitly marked deprecated here
```
