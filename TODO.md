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

## `delve` debugging
https://github.com/derekparker/delve  
https://blog.gopheracademy.com/advent-2015/debugging-with-delve/
  - [ ] Debugging use `delve`
  - [ ] like Microsoft vs-code feature
  - [ ] Set breakpoint with `sign` and key mapping
  - [ ] `lldb.nvim` like Debugging UI

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
  - [ ] gotests (`GoGenerateTestFunc`)
    - https://github.com/cweill/gotests

## Unit Test
  - [ ] Use `go test` feature


# vim-go compatible

- [ ] [GoAlternate](#goalternate)
- [ ] [GoBuild](#gobuild)
- [ ] [GoCoverage](#gocoverage)
- [ ] [GoInfo](#goinfo)
- [ ] [GoInstall](#goinstall)
- [ ] [GoLint](#golint)
- [ ] [GoTest](#gotest)
- [ ] [GoGuru](#goguru)

## GoAlternate
https://github.com/fatih/vim-go/blob/master/autoload/go/alternate.vim
  - [ ] Implements `GoAlternate` command

## GoBuild
https://github.com/fatih/vim-go/blob/master/autoload/go/cmd.vim#L16
  - [ ] Fix display the wrong file path to the `quickfix` or `location-list`
    - `getcwd()`? `expand('%:p')`? or other?
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

## GoLint, lint tools
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
  - [ ] Implements `GoTest` command output to neovim terminal feature
  - [ ] Support `run=func` flag
  - [ ] Support GoTestCompile(?)

## GoGuru
  - [x] Support unsaved file (buffer)
  - [ ] `definition` subcommand support use cgo file (need fix `guru` core)
  - [ ] Implements tags flag feature
  - [ ] Support stacking
