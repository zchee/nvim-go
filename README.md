# nvim-go

[![CircleCI][circleci-badge]][circleci] [![codecov.io][codecov-badge]][codecov] [![godoc.org][godoc-badge]][godoc] [![Releases][release-badge]][release] [![GA][ga-badge]][ga]

nvim-go is a Go development plugin for Neovim written in **pure** Go.

## Requirements

### Neovim

[Installing Neovim - Neovim wiki](https://github.com/neovim/neovim/wiki/Installing-Neovim)

### Go

[Getting Started - The Go Programming Language](https://golang.org/doc/install)

Requires Go `1.11.x` or higter. or, use `devel`.


## Install

nvim-go uses [Go 1.11 Modules](https://github.com/golang/go/wiki/Modules).

We can build nvim-go outside `$GOPATH`. It's still early development feature. For use it, needs to:

```sh
export GO111MODULE='on'
```

After that, Just add following line to your `init.vim`:

```vim
" dein.vim
call dein#add('zchee/nvim-go', {'build': 'make'})

" NeoBundle
NeoBundle 'zchee/nvim-go', {'build': {'unix': 'make'}}

" vim-plug
Plug 'zchee/nvim-go', { 'do': 'make'}
```

## Features

- [ ] First goal is fully compatible vim-go.
  - See [TODO.md](docs/TODO.md#vim-go-compatible).
- [ ] Delve debugger GUI interface.

## Acknowledgement

- [fatih/vim-go](https://github.com/fatih/vim-go)
  - nvim-go is largely inspired by vim-go. Thanks [@fatih](https://github.com/fatih) and vim-go's [contributors](https://github.com/fatih/vim-go/graphs/contributors).
- [neovim/go-client](https://github.com/neovim/go-client)
  - Official Go client for Neovim remote plugin interface.
  - The first architecture was written by [@garyburd](https://github.com/garyburd).
- Authors of vendor packages.
- The Go Authors.

## Donation

Please donate to the location in need of donations in **your country**.

Peace on Earth.

## License

nvim-go is released under the BSD 3-Clause License.


<!-- badge links -->
[circleci]: https://circleci.com/gh/zchee/workflows/nvim-go
[codecov]: https://codecov.io/gh/zchee/nvim-go
[godoc]: https://godoc.org/github.com/zchee/nvim-go
[release]: https://github.com/zchee/nvim-go/releases
[ga]: https://github.com/zchee/nvim-go

[circleci-badge]: https://img.shields.io/circleci/project/github/zchee/nvim-go/master.svg?style=for-the-badge&label=CIRCLECI&logoWidth=16&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgdmlld0JveD0iMCAwIDIwMCAyMDAiPjxwYXRoIGZpbGw9IiNEREQiIGQ9Ik03NC43IDEwMGMwLTEzLjIgMTAuNy0yMy44IDIzLjgtMjMuOCAxMy4xIDAgMjMuOCAxMC43IDIzLjggMjMuOCAwIDEzLjEtMTAuNyAyMy44LTIzLjggMjMuOC0xMy4xIDAtMjMuOC0xMC43LTIzLjgtMjMuOHpNOTguNSAwQzUxLjggMCAxMi43IDMyIDEuNiA3NS4yYy0uMS4zLS4xLjYtLjEgMSAwIDIuNiAyLjEgNC44IDQuOCA0LjhoNDAuM2MxLjkgMCAzLjYtMS4xIDQuMy0yLjggOC4zLTE4IDI2LjUtMzAuNiA0Ny42LTMwLjYgMjguOSAwIDUyLjQgMjMuNSA1Mi40IDUyLjRzLTIzLjUgNTIuNC01Mi40IDUyLjRjLTIxLjEgMC0zOS4zLTEyLjUtNDcuNi0zMC42LS44LTEuNi0yLjQtMi44LTQuMy0yLjhINi4zYy0yLjYgMC00LjggMi4xLTQuOCA0LjggMCAuMy4xLjYuMSAxQzEyLjYgMTY4IDUxLjggMjAwIDk4LjUgMjAwYzU1LjIgMCAxMDAtNDQuOCAxMDAtMTAwUzE1My43IDAgOTguNSAweiIvPjwvc3ZnPg%3D%3D
[codecov-badge]: https://img.shields.io/codecov/c/github/zchee/nvim-go.svg?style=for-the-badge&label=CODECOV&logoWidth=16&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCIgdmlld0JveD0iMCAwIDQ1IDQ1Ij48cGF0aCBkPSJNMjMuMDEzIDBDMTAuMzMzLjAwOS4wMSAxMC4yMiAwIDIyLjc2MnYuMDU4bDMuOTE0IDIuMjc1LjA1My0uMDM2YTExLjI5MSAxMS4yOTEgMCAwIDEgOC4zNTItMS43NjcgMTAuOTExIDEwLjkxMSAwIDAgMSA1LjUgMi43MjZsLjY3My42MjQuMzgtLjgyOGMuMzY4LS44MDIuNzkzLTEuNTU2IDEuMjY0LTIuMjQuMTktLjI3Ni4zOTgtLjU1NC42MzctLjg1MWwuMzkzLS40OS0uNDg0LS40MDRhMTYuMDggMTYuMDggMCAwIDAtNy40NTMtMy40NjYgMTYuNDgyIDE2LjQ4MiAwIDAgMC03LjcwNS40NDlDNy4zODYgMTAuNjgzIDE0LjU2IDUuMDE2IDIzLjAzIDUuMDFjNC43NzkgMCA5LjI3MiAxLjg0IDEyLjY1MSA1LjE4IDIuNDEgMi4zODIgNC4wNjkgNS4zNSA0LjgwNyA4LjU5MWExNi41MyAxNi41MyAwIDAgMC00Ljc5Mi0uNzIzbC0uMjkyLS4wMDJhMTYuNzA3IDE2LjcwNyAwIDAgMC0xLjkwMi4xNGwtLjA4LjAxMmMtLjI4LjAzNy0uNTI0LjA3NC0uNzQ4LjExNS0uMTEuMDE5LS4yMTguMDQxLS4zMjcuMDYzLS4yNTcuMDUyLS41MS4xMDgtLjc1LjE2OWwtLjI2NS4wNjdhMTYuMzkgMTYuMzkgMCAwIDAtLjkyNi4yNzZsLS4wNTYuMDE4Yy0uNjgyLjIzLTEuMzYuNTExLTIuMDE2LjgzOGwtLjA1Mi4wMjZjLS4yOS4xNDUtLjU4NC4zMDUtLjg5OS40OWwtLjA2OS4wNGExNS41OTYgMTUuNTk2IDAgMCAwLTQuMDYxIDMuNDY2bC0uMTQ1LjE3NWMtLjI5LjM2LS41MjEuNjY2LS43MjMuOTYtLjE3LjI0Ny0uMzQuNTEzLS41NTIuODY0bC0uMTE2LjE5OWMtLjE3LjI5Mi0uMzIuNTctLjQ0OS44MjRsLS4wMy4wNTdhMTYuMTE2IDE2LjExNiAwIDAgMC0uODQzIDIuMDI5bC0uMDM0LjEwMmExNS42NSAxNS42NSAwIDAgMC0uNzg2IDUuMTc0bC4wMDMuMjE0YTIxLjUyMyAyMS41MjMgMCAwIDAgLjA0Ljc1NGMuMDA5LjExOS4wMi4yMzcuMDMyLjM1NS4wMTQuMTQ1LjAzMi4yOS4wNDkuNDMybC4wMS4wOGMuMDEuMDY3LjAxNy4xMzMuMDI2LjE5Ny4wMzQuMjQyLjA3NC40OC4xMTkuNzIuNDYzIDIuNDE5IDEuNjIgNC44MzYgMy4zNDUgNi45OWwuMDc4LjA5OC4wOC0uMDk1Yy42ODgtLjgxIDIuMzk1LTMuMzggMi41MzktNC45MjJsLjAwMy0uMDI5LS4wMTQtLjAyNWExMC43MjcgMTAuNzI3IDAgMCAxLTEuMjI2LTQuOTU2YzAtNS43NiA0LjU0NS0xMC41NDQgMTAuMzQzLTEwLjg5bC4zODEtLjAxNGExMS40MDMgMTEuNDAzIDAgMCAxIDYuNjUxIDEuOTU3bC4wNTQuMDM2IDMuODYyLTIuMjM3LjA1LS4wM3YtLjA1NmMuMDA2LTYuMDgtMi4zODQtMTEuNzkzLTYuNzI5LTE2LjA4OUMzNC45MzIgMi4zNjEgMjkuMTYgMCAyMy4wMTMgMCIgZmlsbD0iI0YwMUY3QSIgZmlsbC1ydWxlPSJldmVub2RkIi8+PC9zdmc+
[godoc-badge]: https://img.shields.io/badge/godoc-reference-4F73B3.svg?style=for-the-badge&label=GODOC.ORG&logoWidth=25&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCIgdmlld0JveD0iODUgNTUgMTIwIDEyMCI+PHBhdGggZmlsbD0iIzJEQkNBRiIgZD0iTTQwLjIgMTAxLjFjLS40IDAtLjUtLjItLjMtLjVsMi4xLTIuN2MuMi0uMy43LS41IDEuMS0uNWgzNS43Yy40IDAgLjUuMy4zLjZsLTEuNyAyLjZjLS4yLjMtLjcuNi0xIC42bC0zNi4yLS4xek0yNS4xIDExMC4zYy0uNCAwLS41LS4yLS4zLS41bDIuMS0yLjdjLjItLjMuNy0uNSAxLjEtLjVoNDUuNmMuNCAwIC42LjMuNS42bC0uOCAyLjRjLS4xLjQtLjUuNi0uOS42bC00Ny4zLjF6TTQ5LjMgMTE5LjVjLS40IDAtLjUtLjMtLjMtLjZsMS40LTIuNWMuMi0uMy42LS42IDEtLjZoMjBjLjQgMCAuNi4zLjYuN2wtLjIgMi40YzAgLjQtLjQuNy0uNy43bC0yMS44LS4xek0xNTMuMSA5OS4zYy02LjMgMS42LTEwLjYgMi44LTE2LjggNC40LTEuNS40LTEuNi41LTIuOS0xLTEuNS0xLjctMi42LTIuOC00LjctMy44LTYuMy0zLjEtMTIuNC0yLjItMTguMSAxLjUtNi44IDQuNC0xMC4zIDEwLjktMTAuMiAxOSAuMSA4IDUuNiAxNC42IDEzLjUgMTUuNyA2LjguOSAxMi41LTEuNSAxNy02LjYuOS0xLjEgMS43LTIuMyAyLjctMy43aC0xOS4zYy0yLjEgMC0yLjYtMS4zLTEuOS0zIDEuMy0zLjEgMy43LTguMyA1LjEtMTAuOS4zLS42IDEtMS42IDIuNS0xLjZoMzYuNGMtLjIgMi43LS4yIDUuNC0uNiA4LjEtMS4xIDcuMi0zLjggMTMuOC04LjIgMTkuNi03LjIgOS41LTE2LjYgMTUuNC0yOC41IDE3LTkuOCAxLjMtMTguOS0uNi0yNi45LTYuNi03LjQtNS42LTExLjYtMTMtMTIuNy0yMi4yLTEuMy0xMC45IDEuOS0yMC43IDguNS0yOS4zIDcuMS05LjMgMTYuNS0xNS4yIDI4LTE3LjMgOS40LTEuNyAxOC40LS42IDI2LjUgNC45IDUuMyAzLjUgOS4xIDguMyAxMS42IDE0LjEuNi45LjIgMS40LTEgMS43eiIvPjxwYXRoIGZpbGw9IiMyREJDQUYiIGQ9Ik0xODYuMiAxNTQuNmMtOS4xLS4yLTE3LjQtMi44LTI0LjQtOC44LTUuOS01LjEtOS42LTExLjYtMTAuOC0xOS4zLTEuOC0xMS4zIDEuMy0yMS4zIDguMS0zMC4yIDcuMy05LjYgMTYuMS0xNC42IDI4LTE2LjcgMTAuMi0xLjggMTkuOC0uOCAyOC41IDUuMSA3LjkgNS40IDEyLjggMTIuNyAxNC4xIDIyLjMgMS43IDEzLjUtMi4yIDI0LjUtMTEuNSAzMy45LTYuNiA2LjctMTQuNyAxMC45LTI0IDEyLjgtMi43LjUtNS40LjYtOCAuOXptMjMuOC00MC40Yy0uMS0xLjMtLjEtMi4zLS4zLTMuMy0xLjgtOS45LTEwLjktMTUuNS0yMC40LTEzLjMtOS4zIDIuMS0xNS4zIDgtMTcuNSAxNy40LTEuOCA3LjggMiAxNS43IDkuMiAxOC45IDUuNSAyLjQgMTEgMi4xIDE2LjMtLjYgNy45LTQuMSAxMi4yLTEwLjUgMTIuNy0xOS4xeiIvPjwvc3ZnPg==
[release-badge]: https://img.shields.io/github/release/zchee/nvim-go.svg?style=for-the-badge
[ga-badge]: https://gh-ga-beacon.appspot.com/UA-89201129-1/zchee/nvim-go?flat&useReferer&pixel
