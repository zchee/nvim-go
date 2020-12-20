# nvim-go

[![CircleCI][circleci-badge]][circleci] [![codecov.io][codecov-badge]][codecov] [![pkg.go.dev][pkg.go.dev-badge]][pkg.go.dev] [![Releases][release-badge]][release] [![GA][ga-badge]][ga]

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
[circleci]: https://app.circleci.com/pipelines/github/zchee/nvim-go
[codecov]: https://codecov.io/gh/zchee/nvim-go/branch/main
[pkg.go.dev]: https://pkg.go.dev/github.com/zchee/nvim-go
[release]: https://github.com/zchee/nvim-go/releases
[ga]: https://github.com/zchee/nvim-go

[circleci-badge]: https://img.shields.io/circleci/build/github/zchee/nvim-go/main.svg?logo=circleci&label=circleci&style=for-the-badge
[codecov-badge]: https://img.shields.io/codecov/c/github/zchee/nvim-go/main?logo=codecov&style=for-the-badge
[pkg.go.dev-badge]: https://bit.ly/pkg-go-dev-badge
[release-badge]: https://img.shields.io/github/release/zchee/nvim-go.svg?style=for-the-badge
[ga-badge]: https://gh-ga-beacon.appspot.com/UA-89201129-1/zchee/nvim-go?flat&useReferer&pixel
