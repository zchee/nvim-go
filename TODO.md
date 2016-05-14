# nvim-go original behavior

## Compile error
  - Commands
      - [ ] `GoBuild`
      - [ ] `GoCoverage`
      - [ ] `GoInstall`
      - [ ] `GoTest`
      - [ ] `GoLint`
  - [ ] Implements highlight `sign` to error & warning (like YCM, vim-flake8)
  - [ ] Implements `echo` error & warning message when move cursor to this line

## `GoAstView`
  - [ ] Goal is easy to analysis for Go sources
  - [ ] Alternative tagbar feature
  - [ ] Support display the current cursor `<cword>` AST
  - [ ] Support jump to child AST with any key-mapping
  - [ ] Support tagbar like jump to `func, type, var, const` source position with `<CR>` mapping

## `GoWatch`
  - [ ] Implements `GoWatch` command
  - [ ] Watch the `*.go`, `*.c` and other cgo files in the current package and automatically real build
  - [ ] Use `inotify` for Linux, `fsevents` for OS X
      - [ ] Create `go-notify` package?
  - [ ] Show build and watch log in the split buffer

## AST based syntax highlighting
  - [ ] Re-highlighting color syntax for current buffer based by AST information
    - [ ] Ref: https://github.com/myitcv/neogo/blob/master/neogo.go

## `Dlv`
`delve` debugging

https://github.com/derekparker/delve  
https://blog.gopheracademy.com/advent-2015/debugging-with-delve/

  - [x] Debugging use `delve`
  - [ ] `lldb.nvim` like Debugging UI
  - [ ] Ref: Microsoft vs-code feature
      - https://github.com/Microsoft/vscode-go
  - [ ] Ref: go-debug - go debugger for atom
      - https://github.com/lloiser/go-debug
  - [ ] Set breakpoint with `sign` and key mapping

## `lldb` debugging
http://ribrdb.github.io/lldb/
  - [ ] Debugging use lldb for cgo and more low level debug
  - [ ] Use lldb bindings for Go
      - [ ] Will create `go-lldb` package
  - [ ] Set breakpoint with `sign` and key mapping

## Full cgo support
  - [ ] Support all cgo feature
      - [ ] Use go-clang: https://github.com/go-clang/v3.7
      - [ ] Go cgo internal sources:
          - https://github.com/golang/go/tree/master/src/cmd/cgo
          - https://github.com/golang/go/tree/master/src/runtime/cgo
  - [ ] Definition(Jump to) `C.` func or var source
  - [x] cgo completion was Implemented `deoplete-go` use libclang-python3

## Support useful gotools
  - [x] gotests (`GoGenerateTest`)
      - https://github.com/cweill/gotests

## Unit Test
  - [ ] Use `go test` feature


# vim-go compatible

- [x] [GoAlternate](#goalternate---gotestswitch) -> [GoTestSwitch](#goalternate---gotestswitch)
- [ ] [GoBuild](#gobuild)
- [ ] [GoCoverage](#gocoverage)
- [ ] [GoInfo](#goinfo)
- [ ] [GoInstall](#goinstall)
- [ ] [GoLint](#golint)
- [ ] [GoTest](#gotest)
- [ ] [GoGuru](#goguru)

## GoAlternate -> GoTestSwitch
https://github.com/fatih/vim-go/blob/master/autoload/go/alternate.vim
  - [x] Implements `GoTestSwitch` command
      - [x] Jump to the corresponding (test)function based by parses the AST information
      - [x] Instead of `GoAlternate`

## GoBuild 
https://github.com/fatih/vim-go/blob/master/autoload/go/cmd.vim#L16
  - [x] Fix display the wrong file path to the `quickfix` or `location-list`
      - [ ] Fixed but less than perfect
  - [ ] Inline build(no spawn `go build`) if possible

## GoCoverage
https://github.com/fatih/vim-go/blob/master/autoload/go/cmd.vim
  - [ ] Implements `GoCoverage` command
      - [ ] `go test -coverprofile`
  - [ ] Support other coverage tools
      - [ ] goveralls: https://github.com/mattn/goveralls

## GoInfo
https://github.com/fatih/vim-go/blob/master/autoload/go/complete.vim#L99
  - [ ] Implements `GoInfo` command use guru
  - [ ] Support timer without vim's `updatetime` value
  - [ ] Do not re-call if same code on current cursor

## GoInstall
https://github.com/fatih/vim-go/blob/master/autoload/go/cmd.vim#L145
  - [ ] Implements `GoInstall` command

## GoLint and other lint tools
https://github.com/fatih/vim-go/blob/master/autoload/go/lint.vim
  - [ ] Goal is full analysis to Go sources and lint, like `flake8` tool
      - [ ] Will create yet another `gometalinter` tool from scratch if necessary
      - [ ] `flake8` is defact standard in Python, Refer to `flake8` internal and plugin interface
  - [ ] Implements `golint` only command (`GoLint`)
  - [ ] Implements `govet` only command (`GoVet`)
  - [ ] Support other linter tools
      - [ ] errcheck: https://github.com/kisielk/errcheck
      - [ ] lll: https://github.com/walle/lll

## GoTest
https://github.com/fatih/vim-go/blob/master/autoload/go/cmd.vim#L188
  - [x] Implements `GoTest` command output to neovim terminal feature
  - [ ] Support `run=func` flag
  - [ ] Support GoTestCompile(?)

## GoGuru
  - [x] Support unsaved file (buffer)
  - [ ] `definition` subcommand support use cgo file (need fix `guru` core)
      - [x] Tentatively workaround: https://github.com/zchee/nvim-go/commit/950aa062bd0e7086de3c11753e1bc4ea083e6334
      - [ ] Less than perfect. Maybe can't parse the `struct` provided behavior
  - [ ] Implements tags flag feature
  - [ ] Support stacking


# Command diff list

| Done                   | vim-go commands     | vim-go functions                                    | nvim-go                     | async     |
|:----------------------:|---------------------|-----------------------------------------------------|-----------------------------|:---------:|
| <ul><li>[ ] </li></ul> | `GoInstallBinaries` | `s:GoInstallBinaries(-1)`                           | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoUpdateBinaries`  | `s:GoInstallBinaries(1)`                            | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoPath`            | `go#path#GoPath(<f-args>)`                          | -                           | -         |
| <ul><li>[x] </li></ul> | `GoRename`          | `go#rename#Rename(<bang>0,<f-args>)`                | `Gorename`                  | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoGuruScope`       | `go#guru#Scope(<f-args>)`                           | -                           | -         |
| <ul><li>[x] </li></ul> | `GoImplements`      | `go#guru#Implements(<count>)`                       | `GoGuruImplements`          | **Yes**   |
| <ul><li>[x] </li></ul> | `GoCallees`         | `go#guru#Callees(<count>)`                          | `GoGuruCallees`             | **Yes**   |
| <ul><li>[x] </li></ul> | `GoDescribe`        | `go#guru#Describe(<count>)`                         | `GoGuruDescribe`            | **Yes**   |
| <ul><li>[x] </li></ul> | `GoCallers`         | `go#guru#Callers(<count>)`                          | `GoGuruCallers`             | **Yes**   |
| <ul><li>[x] </li></ul> | `GoCallstack`       | `go#guru#Callstack(<count>)`                        | `GoGuruCallstack`           | **Yes**   |
| <ul><li>[x] </li></ul> | `GoFreevars`        | `go#guru#Freevars(<count>)`                         | `GoGuruFreevars`            | **Yes**   |
| <ul><li>[x] </li></ul> | `GoChannelPeers`    | `go#guru#ChannelPeers(<count>)`                     | `GoGuruChannelPeers`        | **Yes**   |
| <ul><li>[x] </li></ul> | `GoReferrers`       | `go#guru#Referrers(<count>)`                        | `GoGuruReferrers`           | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoGuruTags`        | `go#guru#Tags(<f-args>)`                            | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoSameIds`         | `go#guru#SameIds(<count>)`                          | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoFiles`           | `go#tool#Files()`                                   | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDeps`            | `go#tool#Deps()`                                    | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoInfo`            | `go#complete#Info(0)`                               | -                           | -         |
| <ul><li>[x] </li></ul> | `GoBuild`           | `go#cmd#Build(<bang>0,<f-args>)`                    | `Gobuild`                   | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoGenerate`        | `go#cmd#Generate(<bang>0,<f-args>)`                 | -                           | -         |
| <ul><li>[x] </li></ul> | `GoRun`             | `go#cmd#Run(<bang>0,<f-args>)`                      | `Gorun`                     | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoInstall`         | `go#cmd#Install(<bang>0, <f-args>)`                 | -                           | -         |
| <ul><li>[x] </li></ul> | `GoTest`            | `go#cmd#Test(<bang>0, 0, <f-args>)`                 | `Gotest`                    | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoTestFunc`        | `go#cmd#TestFunc(<bang>0, <f-args>)`                | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoTestCompile`     | `go#cmd#Test(<bang>0, 1, <f-args>)`                 | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoCoverage`        | `go#coverage#Buffer(<bang>0, <f-args>)`             | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoCoverageClear`   | `go#coverage#Clear()`                               | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoCoverageToggle`  | `go#coverage#BufferToggle(<bang>0, <f-args>)`       | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoCoverageBrowser` | `go#coverage#Browser(<bang>0, <f-args>)`            | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoPlay`            | `go#play#Share(<count>, <line1>, <line2>)`          | -                           | -         |
| <ul><li>[x] </li></ul> | `GoDef`             | `go#def#Jump('')`                                   | `call GoGuru('definition')` | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoDefPop`          | `go#def#StackPop(<f-args>)`                         | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDefStack`        | `go#def#Stack(<f-args>)`                            | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDefStackClear`   | `go#def#StackClear(<f-args>)`                       | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDoc`             | `go#doc#Open('new', 'split', <f-args>)`             | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDocBrowser`      | `go#doc#OpenBrowser(<f-args>)`                      | -                           | -         |
| <ul><li>[x] </li></ul> | `GoFmt`             | `go#fmt#Format(-1)`                                 | `Gofmt`                     | ***Any*** |
| <ul><li>[x] </li></ul> | `GoImports`         | `go#fmt#Format(1)`                                  | `Gofmt`                     | ***Any*** |
| <ul><li>[ ] </li></ul> | `GoDrop`            | `go#import#SwitchImport(0, '', <f-args>, '')`       | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoImport`          | `go#import#SwitchImport(1, '', <f-args>, '<bang>')` | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoImportAs`        | `go#import#SwitchImport(1, <f-args>, '<bang>')`     | -                           | -         |
| <ul><li>[x] </li></ul> | `GoMetaLinter`      | `go#lint#Gometa(0, <f-args>)`                       | `Gometalinter`              | **Yes**   |
| <ul><li>[ ] </li></ul> | `GoLint`            | `go#lint#Golint(<f-args>)`                          | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoVet`             | `go#lint#Vet(<bang>0, <f-args>)`                    | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoErrCheck`        | `go#lint#Errcheck(<f-args>)`                        | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoAlternate`       | `go#alternate#Switch(<bang>0, '')`                  | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDecls`           | `ctrlp#init(ctrlp#decls#cmd(0, <q-args>))`          | -                           | -         |
| <ul><li>[ ] </li></ul> | `GoDeclsDir`        | `ctrlp#init(ctrlp#decls#cmd(1, <q-args>))`          | -                           | -         |
