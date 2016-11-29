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
