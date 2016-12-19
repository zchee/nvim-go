# Go typical errors

## initialization loop

- [x] Support includeing all error messages, not first line only

```sh
# github.com/zchee/appleopensource/cmd/gaos
cmd/gaos/versions.go:14: initialization loop:
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:14 cmdVersions refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:19 runVersions refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:16 versionsPkg refers to
        /Users/zchee/go/src/github.com/zchee/appleopensource/cmd/gaos/versions.go:14 cmdVersions
```

## cgo side error

- [ ] Disable C/C++(actually clang, not tested gcc) compiler warning, or ignore some errors if `filepath.Ext(...) == ""`
 - reproduced using [go-clang/v3.9](https://github.com/go-clang/v3.9)

### error

```sh
# github.com/zchee/clang-server/vendor/github.com/go-clang/v3.9/clang
cgo-gcc-prolog:244:6: warning: 'clang_getDiagnosticCategoryName' is deprecated [-Wdeprecated-declarations]
../vendor/github.com/go-clang/v3.9/clang/clang-c/Index.h:952:10: note: 'clang_getDiagnosticCategoryName' has been explicitly marked deprecated here
```

### quickfix

```vim
github.com/zchee/clang-server/vendor/github.com/go-clang/v3.9/clang/cgo-gcc-prolog|244 col 6| warning: 'clang_getDiagnosticCategoryName' is deprecated [-Wdeprecated-declarations]
../vendor/github.com/go-clang/v3.9/clang/clang-c/Index.h|952 col 10| note: 'clang_getDiagnosticCategoryName' has been explicitly marked deprecated here
```

## have...want Go compiler type suggestion

- [ ] Need suuport including have (...) and want (...) suggest to quickfix

### error

```sh
# github.com/zchee/clang-server/compilationdatabase
./compilationdatabase.go:99: too many arguments to return
  have ([]string, nil)
  want (error)
./compilationdatabase.go:140: undefined: filename
./compilationdatabase.go:141: undefined: filename
./compilationdatabase.go:141: too many arguments to return
  have (<T>, nil)
  want (error)
./compilationdatabase.go:144: undefined: filename
./compilationdatabase.go:145: no new variables on left side of :=
./compilationdatabase.go:161: too many arguments to return
  have (nil, error)
  want (error)
./compilationdatabase.go:167: too many arguments to return
  have (nil, error)
  want (error)
./compilationdatabase.go:177: no new variables on left side of :=
./compilationdatabase.go:178: no new variables on left side of :=
```

### quickfix

```vim
compilationdatabase.go|99| too many arguments to return
compilationdatabase.go|140| undefined: filename
compilationdatabase.go|141| undefined: filename
compilationdatabase.go|141| too many arguments to return
compilationdatabase.go|144| undefined: filename
compilationdatabase.go|145| no new variables on left side of :=
compilationdatabase.go|161| too many arguments to return
compilationdatabase.go|167| too many arguments to return
```

## cannot find package

### error

```sh
indexdb/indexdb.go:14:2: cannot find package "github.com/zchee/clang-server/symbol" in any of:
	/Users/zchee/go/src/github.com/zchee/clang-server/vendor/github.com/zchee/clang-server/symbol (vendor tree)
	/usr/local/go/src/github.com/zchee/clang-server/symbol (from $GOROOT)
	/Users/zchee/go/src/github.com/zchee/clang-server/symbol (from $GOPATH)
```

### quickfix

```vim
indexdb.go|14 col 2| cannot find package "github.com/zchee/clang-server/symbol" in any of:
```
