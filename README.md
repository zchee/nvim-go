nvim-go
=======

[![CircleCI][circleci-badge]][circleci] [![codecov.io][codecov-badge]][codecov] [![Releases][release-badge]][release] [![GA][ga-badge]][ga]

Go development plugin for Neovim written in pure Go.

Requirements
------------

### Neovim

[Installing Neovim - Neovim wiki](https://github.com/neovim/neovim/wiki/Installing-Neovim)

### Go

[Getting Started - The Go Programming Language](https://golang.org/doc/install)


Install
-------

nvim-go do not support `go get` install, because Neovim's runtimepath is not under the `$GOPATH`.  
Currently, depends on the [constabulary/gb](https://github.com/constabulary/gb).

```sh
go get -u github.com/constabulary/gb/...
```

After installed gb, add your init.vim:

```vim
" dein.vim
call dein#add('zchee/nvim-go', {'build': 'make'})

" NeoBundle
NeoBundle 'zchee/nvim-go', {'build': {'unix': 'make'}}

" vim-plug
Plug 'zchee/nvim-go', { 'do': 'make'}
```

Features
--------

- First goal is fully compatible vim-go. See [TODO.md](docs/TODO.md#vim-go-compatible).
- Delve debugger GUI interface.

Acknowledgement
---------------

- [fatih/vim-go](https://github.com/fatih/vim-go)
  - nvim-go is largely inspired by vim-go. Thanks [@fatih](https://github.com/fatih) and vim-go's [contributors](https://github.com/fatih/vim-go/graphs/contributors).
- [neovim/go-client](https://github.com/neovim/go-client)
  - Official Go client for Neovim remote plugin interface. written by [@garyburd](https://github.com/garyburd)
- Authors of vendor packages.
- The Go Authors.

Donation
--------

Please donate to the location in need of donations in your country. peace.

License
-------

nvim-go is released under the BSD 3-Clause License.


[circleci-badge]: https://img.shields.io/circleci/project/github/zchee/nvim-go.svg?style=flat-square
[circleci]: https://circleci.com/gh/zchee/nvim-go
[codecov-badge]: https://img.shields.io/codecov/c/github/zchee/nvim-go.svg?style=flat-square
[codecov]: https://codecov.io/gh/zchee/nvim-go
[release-badge]: https://img.shields.io/github/release/zchee/nvim-go.svg?style=flat-square
[release]: https://github.com/zchee/nvim-go/releases
[ga-badge]: https://ga-beacon.appspot.com/UA-89201129-1/nvim-go?flat&useReferer&pixel
[ga]: https://github.com/zchee/nvim-go
