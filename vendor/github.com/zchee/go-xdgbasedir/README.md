# go-xdgbasedir

[![XDGBaseDirSpecVersion](https://img.shields.io/badge/%20XDG%20Base%20Dir%20-%200.8-blue.svg?style=flat-square)](https://specifications.freedesktop.org/basedir-spec/0.8/)

Package `xdgbasedir` implements a freedesktop XDG Base Directory Specification for Go.

| **linux / darwin**                          | **windows**                                   | **coverage**                            |
|:-------------------------------------------:|:---------------------------------------------:|:---------------------------------------:|
| [![circleci.com][circleci-badge]][circleci] | [![appveyor-badge][appveyor-badge]][appveyor] | [![codecov.io][codecov-badge]][codecov] |
| **godoc.org**                               | **Release**                                   |                                         |
| [![godoc.org][godoc-badge]][godoc]          | [![Release][release-badge]][release]          |                                         |

## freedesktop.org XDG Base Directory Specification reference:

  - https://specifications.freedesktop.org/basedir-spec/basedir-spec-0.8.html
  - https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

## Current specification version:
  
  - 0.8

## Specification

| func           | linux, darwin (Mode: `Unix`)  | darwin (Mode: `Native`)         | windows                               |
|----------------|-------------------------------|---------------------------------|---------------------------------------|
| `DataHome()`   | `~/.local/share`              | `~/Library/Application Support` | `C:\Users\%USER%\AppData\Local`       |
| `ConfigHome()` | `~/.config`                   | `~/Library/Preferences`         | `C:\Users\%USER%\AppData\Local`       |
| `DataDirs()`   | `/usr/local/share:/usr/share` | `~/Library/Application Support` | `C:\Users\%USER%\AppData\Local`       |
| `ConfigDirs()` | `/etc/xdg`                    | `~/Library/Preferences`         | `C:\Users\%USER%\AppData\Local`       |
| `CacheHome()`  | `~/.cache`                    | `~/Library/Caches`              | `C:\Users\%USER%\AppData\Local\cache` |
| `RuntimeDir()` | `/run/user/$(id -u)`          | `~/Library/Application Support` | `C:\Users\%USER%`                     |

## Note

XDG Base Directory Specification is mainly for GNU/Linux. It does not mention which directory to use with macOS(`darwin`) or `windows`.  
So, We referred to the [qt standard paths document](http://doc.qt.io/qt-5/qstandardpaths.html) for the corresponding directory.

We prepared a `Mode` for users using macOS like Unix. It's `darwin` GOOS specific.  
If it is set to `Unix`, it refers to the same path as linux. If it is set to `Native`, it refers to the [Specification](#specification) path.  
By default, `Unix`.

`Unix`:

```go
// +build darwin

package main

import (
	"fmt"

	"github.com/zchee/go-xdgbasedir"
)

func main() {
	xdgbasedir.Mode = xdgbasedir.Unix // optional, default is Unix
	fmt.Println(xdgbasedir.DataHome())

	// Output:
	// "/Users/foo/.local/share"
}
```

`Native`:

```go
// +build darwin

package main

import (
	"fmt"

	"github.com/zchee/go-xdgbasedir"
)

func main() {
	xdgbasedir.Mode = xdgbasedir.Native
	fmt.Println(xdgbasedir.DataHome())

	// Output:
	// "/Users/foo/Library/Application Support"
}
```

## Badge

powered by [shields.io](https://shields.io).

[![XDGBaseDirSpecVersion](https://img.shields.io/badge/%20XDG%20Base%20Dir%20-%200.8-blue.svg?style=flat-square)](https://specifications.freedesktop.org/basedir-spec/0.8/)

markdown:
```markdown
[![XDGBaseDirSpecVersion](https://img.shields.io/badge/%20XDG%20Base%20Dir%20-%200.8-blue.svg?style=flat-square)](https://specifications.freedesktop.org/basedir-spec/0.8/)
```

<!-- badge links -->
[circleci]: https://circleci.com/gh/zchee/workflows/go-xdgbasedir
[appveyor]: https://ci.appveyor.com/project/zchee/go-xdgbasedir
[godoc]: https://godoc.org/github.com/zchee/go-xdgbasedir
[codecov]: https://codecov.io/gh/zchee/go-xdgbasedir
[release]: https://github.com/zchee/go-xdgbasedir/releases

[circleci-badge]: https://img.shields.io/circleci/project/github/zchee/go-xdgbasedir/master.svg?style=flat-square&label=%20%20CircleCI&logoWidth=16&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgdmlld0JveD0iMCAwIDIwMCAyMDAiPjxwYXRoIGZpbGw9IiNEREQiIGQ9Ik03NC43IDEwMGMwLTEzLjIgMTAuNy0yMy44IDIzLjgtMjMuOCAxMy4xIDAgMjMuOCAxMC43IDIzLjggMjMuOCAwIDEzLjEtMTAuNyAyMy44LTIzLjggMjMuOC0xMy4xIDAtMjMuOC0xMC43LTIzLjgtMjMuOHpNOTguNSAwQzUxLjggMCAxMi43IDMyIDEuNiA3NS4yYy0uMS4zLS4xLjYtLjEgMSAwIDIuNiAyLjEgNC44IDQuOCA0LjhoNDAuM2MxLjkgMCAzLjYtMS4xIDQuMy0yLjggOC4zLTE4IDI2LjUtMzAuNiA0Ny42LTMwLjYgMjguOSAwIDUyLjQgMjMuNSA1Mi40IDUyLjRzLTIzLjUgNTIuNC01Mi40IDUyLjRjLTIxLjEgMC0zOS4zLTEyLjUtNDcuNi0zMC42LS44LTEuNi0yLjQtMi44LTQuMy0yLjhINi4zYy0yLjYgMC00LjggMi4xLTQuOCA0LjggMCAuMy4xLjYuMSAxQzEyLjYgMTY4IDUxLjggMjAwIDk4LjUgMjAwYzU1LjIgMCAxMDAtNDQuOCAxMDAtMTAwUzE1My43IDAgOTguNSAweiIvPjwvc3ZnPg%3D%3D
[appveyor-badge]: https://img.shields.io/appveyor/ci/zchee/go-xdgbasedir/master.svg?style=flat-square&label=AppVeyor&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCIgdmlld0JveD0iMCAwIDQwIDQwIj48cGF0aCBmaWxsPSIjREREIiBkPSJNMjAgMGMxMSAwIDIwIDkgMjAgMjBzLTkgMjAtMjAgMjBTMCAzMSAwIDIwIDkgMCAyMCAwem00LjkgMjMuOWMyLjItMi44IDEuOS02LjgtLjktOC45LTIuNy0yLjEtNi43LTEuNi05IDEuMi0yLjIgMi44LTEuOSA2LjguOSA4LjkgMi44IDIuMSA2LjggMS42IDktMS4yem0tMTAuNyAxM2MxLjIuNSAzLjggMSA1LjEgMUwyOCAyNS4zYzIuOC00LjIgMi4xLTkuOS0xLjgtMTMtMy41LTIuOC04LjQtMi43LTExLjkgMEwyLjIgMjEuNmMuMyAzLjIgMS4yIDQuOCAxLjIgNC45bDYuOS03LjVjLS41IDMuMy43IDYuNyAzLjUgOC44IDIuNCAxLjkgNS4zIDIuNCA4LjEgMS44bC03LjcgNy4zeiIvPjwvc3ZnPg%3D%3D
[godoc-badge]: https://img.shields.io/badge/godoc-reference-4F73B3.svg?style=flat-square&label=%20godoc.org
[codecov-badge]: https://img.shields.io/codecov/c/github/zchee/go-xdgbasedir/master.svg?style=flat-square&label=%20%20Codecov%2Eio&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCIgdmlld0JveD0iMCAwIDI1NiAyODEiPjxwYXRoIGZpbGw9IiNFRUUiIGQ9Ik0yMTguNTUxIDM3LjQxOUMxOTQuNDE2IDEzLjI4OSAxNjIuMzMgMCAxMjguMDk3IDAgNTcuNTM3LjA0Ny4wOTEgNTcuNTI3LjA0IDEyOC4xMjFMMCAxNDkuODEzbDE2Ljg1OS0xMS40OWMxMS40NjgtNy44MTQgMjQuNzUtMTEuOTQ0IDM4LjQxNy0xMS45NDQgNC4wNzkgMCA4LjE5OC4zNzMgMTIuMjQgMS4xMSAxMi43NDIgMi4zMiAyNC4xNjUgOC4wODkgMzMuNDE0IDE2Ljc1OCAyLjEyLTQuNjcgNC42MTQtOS4yMDkgNy41Ni0xMy41MzZhODguMDgxIDg4LjA4MSAwIDAgMSAzLjgwNS01LjE1Yy0xMS42NTItOS44NC0yNS42NDktMTYuNDYzLTQwLjkyNi0xOS4yNDVhOTAuMzUgOTAuMzUgMCAwIDAtMTYuMTItMS40NTkgODguMzc3IDg4LjM3NyAwIDAgMC0zMi4yOSA2LjA3YzguMzYtNTEuMjIyIDUyLjg1LTg5LjM3IDEwNS4yMy04OS40MDggMjguMzkyIDAgNTUuMDc4IDExLjA1MyA3NS4xNDkgMzEuMTE3IDE2LjAxMSAxNi4wMSAyNi4yNTQgMzYuMDMzIDI5Ljc4OCA1OC4xMTctMTAuMzI5LTQuMDM1LTIxLjIxMi02LjEtMzIuNDAzLTYuMTQ0bC0xLjU2OC0uMDA3YTkwLjk1NyA5MC45NTcgMCAwIDAtMy40MDEuMTExYy0xLjk1NS4xLTMuODk4LjI3Ny01LjgyMS41LS41NzQuMDYzLTEuMTM5LjE1My0xLjcwNy4yMzEtMS4zNzguMTg2LTIuNzUuMzk1LTQuMTA5LjYzOS0uNjAzLjExLTEuMjAzLjIzMS0xLjguMzUxYTkwLjUxNyA5MC41MTcgMCAwIDAtNC4xMTQuOTM3Yy0uNDkyLjEyNi0uOTgzLjI0My0xLjQ3LjM3NGE5MC4xODMgOTAuMTgzIDAgMCAwLTUuMDkgMS41MzhjLS4xLjAzNS0uMjA0LjA2My0uMzA0LjA5NmE4Ny41MzIgODcuNTMyIDAgMCAwLTExLjA1NyA0LjY0OWMtLjA5Ny4wNS0uMTkzLjEwMS0uMjkzLjE1MWE4Ni43IDg2LjcgMCAwIDAtNC45MTIgMi43MDFsLS4zOTguMjM4YTg2LjA5IDg2LjA5IDAgMCAwLTIyLjMwMiAxOS4yNTNjLS4yNjIuMzE4LS41MjQuNjM1LS43ODQuOTU4LTEuMzc2IDEuNzI1LTIuNzE4IDMuNDktMy45NzYgNS4zMzZhOTEuNDEyIDkxLjQxMiAwIDAgMC0zLjY3MiA1LjkxMyA5MC4yMzUgOTAuMjM1IDAgMCAwLTIuNDk2IDQuNjM4Yy0uMDQ0LjA5LS4wODkuMTc1LS4xMzMuMjY1YTg4Ljc4NiA4OC43ODYgMCAwIDAtNC42MzcgMTEuMjcybC0uMDAyLjAwOXYuMDA0YTg4LjAwNiA4OC4wMDYgMCAwIDAtNC41MDkgMjkuMzEzYy4wMDUuMzk3LjAwNS43OTQuMDE5IDEuMTkyLjAyMS43NzcuMDYgMS41NTcuMTA0IDIuMzM4YTk4LjY2IDk4LjY2IDAgMCAwIC4yODkgMy44MzRjLjA3OC44MDQuMTc0IDEuNjA2LjI3NSAyLjQxLjA2My41MTIuMTE5IDEuMDI2LjE5NSAxLjUzNGE5MC4xMSA5MC4xMSAwIDAgMCAuNjU4IDQuMDFjNC4zMzkgMjIuOTM4IDE3LjI2MSA0Mi45MzcgMzYuMzkgNTYuMzE2bDIuNDQ2IDEuNTY0LjAyLS4wNDhhODguNTcyIDg4LjU3MiAwIDAgMCAzNi4yMzIgMTMuNDVsMS43NDYuMjM2IDEyLjk3NC0yMC44MjItNC42NjQtLjEyN2MtMzUuODk4LS45ODUtNjUuMS0zMS4wMDMtNjUuMS02Ni45MTcgMC0zNS4zNDggMjcuNjI0LTY0LjcwMiA2Mi44NzYtNjYuODI5bDIuMjMtLjA4NWMxNC4yOTItLjM2MiAyOC4zNzIgMy44NTkgNDAuMzI1IDExLjk5N2wxNi43ODEgMTEuNDIxLjAzNi0yMS41OGMuMDI3LTM0LjIxOS0xMy4yNzItNjYuMzc5LTM3LjQ0OS05MC41NTQiLz48L3N2Zz4=
[release-badge]: https://img.shields.io/github/release/zchee/go-xdgbasedir/all.svg?style=flat-square&label=%20%20Release
