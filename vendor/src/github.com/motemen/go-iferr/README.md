# goiferr

## Installation

    go get -u github.com/motemen/go-iferr/cmd/goiferr

## Description

`goiferr` automatically inserts idiomatic error handling to given Go source code.

If there are an assign statement which involves variables with `error` type
and an empty line right after that, `goiferr` inserts `if err != nil { ... }` code there.

The error handling code will be:

- `return nil, err`-like statement if the scope is inside a function whose return types include `error`
- `log.Fatal(err)` if the log package can be referred from the scope
- `t.Fatal(err)` if a variable named `t` whose type of `*testing.T` can be referred from the scope
- `panic(err.Error())` otherwise

See examples for more information.

## Usage

    Usage: goiferr [-w] <args>...
      -w    rewrite input files in place

## Examples

The `return` case:

```diff
 func GoVersion() (string, error) {
        path, err := exec.LookPath("go")
+       if err != nil {
+               return "", err
+       }

        cmd := exec.Command(path, "version")
        b, err := cmd.Output()
+       if err != nil {
+               return "", err
+       }

        return string(b), nil
 }
```

`log` case:

```diff
 import "log"

 func main() {
        _, err := GoVersion()
+       if err != nil {
+               log.Fatal(err)
+       }
+
 }
```

`testing` case:

```diff
 import "testing"

 func TestFoo(t *testing.T) {
        _, err := GoVersion()
+       if err != nil {
+               t.Fatal(err)
+       }
+
 }

```

Otherwise:

```diff
 func main() {
        _, err := GoVersion()
+       if err != nil {
+               panic(err.Error())
+       }
+
 }
```

## TODO

- Generate error handling code correctly inside a function with named return values
- Don't guess the variable name to "err"
- Make error handling code custamizable

## Author

motemen <https://motemen.github.io/>
