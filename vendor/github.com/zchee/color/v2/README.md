# Color

[![CircleCI][circleci-badge]][circleci] [![godoc.org][godoc-badge]][godoc] [![codecov.io][codecov-badge]][codecov] [![Releases][release-badge]][tag] [![GA][ga-badge]][ga]

Color lets you use colorized outputs in terms of [ANSI Escape
Codes](http://en.wikipedia.org/wiki/ANSI_escape_code#Colors) in Go (Golang). It
has support for Windows too! The API can be used in several ways, pick one that
suits you.


![Color](https://i.imgur.com/c1JI0lA.png)


## Install

```sh
go get -u github.com/zchee/color
```

Note that the `vendor` folder is here for stability. Remove the folder if you
already have the dependencies in your GOPATH.

## Examples

### Standard colors

```go
// Print with default helper functions
color.Cyan("Prints text in cyan.")

// A newline will be appended automatically
color.Blue("Prints %s in blue.", "text")

// These are using the default foreground colors
color.Red("We have red")
color.Magenta("And many others ..")

```

### Mix and reuse colors

```go
// Create a new color object
c := color.New(color.FgCyan).Add(color.Underline)
c.Println("Prints cyan text with an underline.")

// Or just add them to New()
d := color.New(color.FgCyan, color.Bold)
d.Printf("This prints bold cyan %s\n", "too!.")

// Mix up foreground and background colors, create new mixes!
red := color.New(color.FgRed)

boldRed := red.Add(color.Bold)
boldRed.Println("This will print text in bold red.")

whiteBackground := red.Add(color.BgWhite)
whiteBackground.Println("Red text with white background.")
```

### Use your own output (io.Writer)

```go
// Use your own io.Writer output
color.New(color.FgBlue).Fprintln(myWriter, "blue color!")

blue := color.New(color.FgBlue)
blue.Fprint(writer, "This will print text in blue.")
```

### Custom print functions (PrintFunc)

```go
// Create a custom print function for convenience
red := color.New(color.FgRed).PrintfFunc()
red("Warning")
red("Error: %s", err)

// Mix up multiple attributes
notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
notice("Don't forget this...")
```

### Custom fprint functions (FprintFunc)

```go
blue := color.New(FgBlue).FprintfFunc()
blue(myWriter, "important notice: %s", stars)

// Mix up with multiple attributes
success := color.New(color.Bold, color.FgGreen).FprintlnFunc()
success(myWriter, "Don't forget this...")
```

### Insert into noncolor strings (SprintFunc)

```go
// Create SprintXxx functions to mix strings with other non-colorized strings:
yellow := color.New(color.FgYellow).SprintFunc()
red := color.New(color.FgRed).SprintFunc()
fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))

info := color.New(color.FgWhite, color.BgGreen).SprintFunc()
fmt.Printf("This %s rocks!\n", info("package"))

// Use helper functions
fmt.Println("This", color.RedString("warning"), "should be not neglected.")
fmt.Printf("%v %v\n", color.GreenString("Info:"), "an important message.")

// Windows supported too! Just don't forget to change the output to color.Output
fmt.Fprintf(color.Output, "Windows support: %s", color.GreenString("PASS"))
```

### Plug into existing code

```go
// Use handy standard colors
color.Set(color.FgYellow)

fmt.Println("Existing text will now be in yellow")
fmt.Printf("This one %s\n", "too")

color.Unset() // Don't forget to unset

// You can mix up parameters
color.Set(color.FgMagenta, color.Bold)
defer color.Unset() // Use it in your function

fmt.Println("All text will now be bold magenta.")
```

### Disable/Enable color

There might be a case where you want to explicitly disable/enable color output. the
`go-isatty` package will automatically disable color output for non-tty output streams
(for example if the output were piped directly to `less`)

`Color` has support to disable/enable colors both globally and for single color
definitions. For example suppose you have a CLI app and a `--no-color` bool flag. You
can easily disable the color output with:

```go

var flagNoColor = flag.Bool("no-color", false, "Disable color output")

if *flagNoColor {
	color.NoColor = true // disables colorized output
}
```

It also has support for single color definitions (local). You can
disable/enable color output on the fly:

```go
c := color.New(color.FgCyan)
c.Println("Prints cyan text")

c.DisableColor()
c.Println("This is printed without any color")

c.EnableColor()
c.Println("This prints again cyan...")
```

## Benchmark

### Run benchmark

```sh
cd ./benchmarks
go mod vendor -v
go test -v -mod=vendor -tags=benchmark -cpu 1,4,12 -count 10 -run='^$' -bench=. -benchtime=.1s . -fatih | tee old.txt
go test -v -mod=vendor -tags=benchmark -cpu 1,4,12 -count 10 -run='^$' -bench=. -benchtime=.1s . | tee new.txt
benchstat old.txt new.txt
```

### Benchmark result

On my Macbook Pro.

- `list_cpu_features` command
  - [google/cpu_features](https://github.com/google/cpu_features)
- `lscpu` command (on macOS)
  - [NanXiao/lscpu](https://github.com/NanXiao/lscpu)

```console
$ system_profiler SPHardwareDataType
# (omitted)
Model Name: MacBook Pro
Model Identifier: MacBookPro15,1
Processor Name: Intel Core i9
Processor Speed: 2.9 GHz
Number of Processors: 1
Total Number of Cores: 6
L2 Cache (per Core): 256 KB
L3 Cache: 12 MB
Memory: 32 GB

$ list_cpu_features
arch            : x86
brand           : Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
family          :   6 (0x06)
model           : 158 (0x9E)
stepping        :  10 (0x0A)
uarch           : INTEL_KBL
flags           : aes,avx,avx2,bmi1,bmi2,cx16,erms,f16c,fma3,movbe,popcnt,rdrnd,sgx,sse4_1,sse4_2,ssse3

$ lscpu
Architecture:            x86_64
Byte Order:              Little Endian
Total CPU(s):            12
Thread(s) per core:      2
Core(s) per socket:      6
Socket(s):               1
Vendor:                  GenuineIntel
CPU family:              6
Model:                   158
Model name:              MacBookPro15,1
Stepping:                10
L1d cache:               32K
L1i cache:               32K
L2 cache:                256K
L3 cache:                12M
Flags:                   fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 cflsh ds acpi mmx fxsr sse sse2 ss htt tm pbe sse3 pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline aes xsave osxsave avx f16c rdrnd syscall nx pdpe1gb rdtscp lm lahf_lm lzcnt
```

```
name                  old time/op    new time/op    delta
NewPrint                18.3µs ± 9%    20.0µs ± 2%   +8.97%  (p=0.000 n=10+9)
NewPrint-4              5.25µs ± 5%    5.34µs ±10%     ~     (p=0.605 n=9+9)
NewPrint-12             3.35µs ±23%    3.62µs ±34%     ~     (p=1.000 n=9+10)
ColorPrint              2.41µs ± 5%    2.13µs ± 3%  -11.76%  (p=0.000 n=10+8)
ColorPrint-4             724ns ±20%     603ns ± 4%  -16.70%  (p=0.000 n=10+8)
ColorPrint-12            526ns ±11%     437ns ± 1%  -16.80%  (p=0.000 n=9+8)
ColorString             3.32µs ± 3%    3.01µs ±19%   -9.28%  (p=0.005 n=9+10)
ColorString-4           1.07µs ±11%    0.99µs ± 4%   -8.11%  (p=0.000 n=9+8)
ColorString-12          1.14µs ±21%    1.01µs ± 5%  -11.35%  (p=0.002 n=10+10)
GetCacheColorFg         49.2ns ±15%    18.2ns ± 5%  -62.99%  (p=0.000 n=10+8)
GetCacheColorFg-4       64.7ns ±12%     6.5ns ±39%  -90.03%  (p=0.000 n=10+10)
GetCacheColorFg-12      99.1ns ± 7%     3.5ns ± 1%  -96.43%  (p=0.000 n=9+10)
GetCacheColorFgHi       48.1ns ±11%    18.1ns ± 7%  -62.45%  (p=0.000 n=9+8)
GetCacheColorFgHi-4     69.9ns ±10%     5.3ns ± 8%  -92.40%  (p=0.000 n=9+8)
GetCacheColorFgHi-12     105ns ± 4%       4ns ± 2%  -96.58%  (p=0.000 n=9+10)
GetCacheColorBg         50.9ns ± 9%    18.5ns ± 6%  -63.58%  (p=0.000 n=10+8)
GetCacheColorBg-4       73.5ns ±10%     6.1ns ±44%  -91.67%  (p=0.000 n=10+10)
GetCacheColorBg-12       106ns ± 6%       4ns ± 1%  -96.61%  (p=0.000 n=8+8)
GetCacheColorBgHi       51.5ns ±11%    19.3ns ±16%  -62.51%  (p=0.000 n=10+9)
GetCacheColorBgHi-4     71.7ns ± 7%     6.5ns ±50%  -90.90%  (p=0.000 n=9+10)
GetCacheColorBgHi-12     106ns ± 9%       4ns ± 4%  -96.56%  (p=0.000 n=10+9)
ColorPrintFg            2.60µs ± 5%    2.31µs ±11%  -11.05%  (p=0.000 n=10+9)
ColorPrintFg-4           711ns ±10%     628ns ± 6%  -11.72%  (p=0.000 n=9+8)
ColorPrintFg-12          529ns ±12%     483ns ±36%   -8.77%  (p=0.029 n=9+9)
ColorPrintFgHi          2.80µs ±11%    2.50µs ±21%  -10.81%  (p=0.009 n=10+10)
ColorPrintFgHi-4         715ns ±18%     633ns ± 6%  -11.46%  (p=0.000 n=9+8)
ColorPrintFgHi-12        566ns ±24%     496ns ±41%  -12.37%  (p=0.021 n=10+9)
ColorPrintBgHi          2.91µs ±12%    2.59µs ±22%  -10.78%  (p=0.011 n=10+10)
ColorPrintBgHi-4         726ns ± 7%     636ns ± 5%  -12.30%  (p=0.000 n=9+9)
ColorPrintBgHi-12        605ns ±24%     556ns ±38%     ~     (p=0.271 n=10+10)
ColorStringFg           3.80µs ±16%    3.41µs ±12%  -10.39%  (p=0.017 n=10+9)
ColorStringFg-4         1.19µs ±16%    1.06µs ± 5%  -10.48%  (p=0.000 n=10+10)
ColorStringFg-12        1.22µs ±14%    1.18µs ± 3%     ~     (p=0.943 n=10+7)
ColorStringFgHi         3.80µs ±11%    3.44µs ±15%   -9.49%  (p=0.011 n=10+10)
ColorStringFgHi-4       1.14µs ±12%    1.05µs ± 3%   -8.25%  (p=0.001 n=10+8)
ColorStringFgHi-12      1.24µs ±13%    1.16µs ± 5%   -5.98%  (p=0.008 n=10+10)
ColorStringBgHi         3.79µs ±23%    3.32µs ±11%  -12.41%  (p=0.006 n=10+9)
ColorStringBgHi-4       1.15µs ±13%    1.06µs ± 8%   -7.99%  (p=0.028 n=10+8)
ColorStringBgHi-12      1.29µs ±17%    1.08µs ± 8%  -16.07%  (p=0.001 n=10+9)

name                  old alloc/op   new alloc/op   delta
NewPrint                 87.0B ± 0%     66.0B ± 0%  -24.14%  (p=0.000 n=10+9)
NewPrint-4               87.4B ± 1%     66.4B ± 2%  -24.03%  (p=0.000 n=10+10)
NewPrint-12              90.9B ± 1%     69.8B ± 3%  -23.21%  (p=0.000 n=10+10)
ColorPrint               85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrint-4             85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrint-12            86.0B ± 0%     64.0B ± 0%  -25.58%  (p=0.000 n=10+10)
ColorString             4.72kB ± 0%    4.69kB ± 0%   -0.78%  (p=0.000 n=9+10)
ColorString-4           4.74kB ± 0%    4.70kB ± 0%   -0.79%  (p=0.000 n=10+10)
ColorString-12          4.77kB ± 0%    4.73kB ± 0%   -0.82%  (p=0.000 n=10+10)
GetCacheColorFg          0.00B          0.00B          ~     (all equal)
GetCacheColorFg-4        0.00B          0.00B          ~     (all equal)
GetCacheColorFg-12       0.00B          0.00B          ~     (all equal)
GetCacheColorFgHi        0.00B          0.00B          ~     (all equal)
GetCacheColorFgHi-4      0.00B          0.00B          ~     (all equal)
GetCacheColorFgHi-12     0.00B          0.00B          ~     (all equal)
GetCacheColorBg          0.00B          0.00B          ~     (all equal)
GetCacheColorBg-4        0.00B          0.00B          ~     (all equal)
GetCacheColorBg-12       0.00B          0.00B          ~     (all equal)
GetCacheColorBgHi        0.00B          0.00B          ~     (all equal)
GetCacheColorBgHi-4      0.00B          0.00B          ~     (all equal)
GetCacheColorBgHi-12     0.00B          0.00B          ~     (all equal)
ColorPrintFg             85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrintFg-4           85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrintFg-12          86.0B ± 0%     64.0B ± 0%  -25.58%  (p=0.000 n=10+10)
ColorPrintFgHi           85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrintFgHi-4         85.0B ± 0%     64.0B ± 0%  -24.71%  (p=0.000 n=10+10)
ColorPrintFgHi-12        86.0B ± 0%     64.0B ± 0%  -25.58%  (p=0.000 n=10+10)
ColorPrintBgHi           90.0B ± 0%     64.0B ± 0%  -28.89%  (p=0.000 n=10+10)
ColorPrintBgHi-4         91.0B ± 0%     64.0B ± 0%  -29.67%  (p=0.000 n=10+10)
ColorPrintBgHi-12        91.0B ± 0%     64.0B ± 0%  -29.67%  (p=0.000 n=9+9)
ColorStringFg           4.72kB ± 0%    4.68kB ± 0%   -0.78%  (p=0.000 n=9+9)
ColorStringFg-4         4.74kB ± 0%    4.70kB ± 0%   -0.79%  (p=0.000 n=7+10)
ColorStringFg-12        4.76kB ± 0%    4.72kB ± 1%   -0.94%  (p=0.000 n=10+10)
ColorStringFgHi         4.72kB ± 0%    4.69kB ± 0%   -0.78%  (p=0.000 n=10+10)
ColorStringFgHi-4       4.74kB ± 0%    4.70kB ± 0%   -0.78%  (p=0.000 n=10+9)
ColorStringFgHi-12      4.77kB ± 0%    4.72kB ± 0%   -0.99%  (p=0.000 n=9+10)
ColorStringBgHi         4.72kB ± 0%    4.69kB ± 0%   -0.75%  (p=0.000 n=8+10)
ColorStringBgHi-4       4.74kB ± 0%    4.70kB ± 0%   -0.80%  (p=0.000 n=9+8)
ColorStringBgHi-12      4.77kB ± 0%    4.73kB ± 0%   -0.69%  (p=0.000 n=10+9)

name                  old allocs/op  new allocs/op  delta
NewPrint                  5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
NewPrint-4                5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
NewPrint-12               5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrint                5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrint-4              5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrint-12             5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorString               9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorString-4             9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorString-12            9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
GetCacheColorFg           0.00           0.00          ~     (all equal)
GetCacheColorFg-4         0.00           0.00          ~     (all equal)
GetCacheColorFg-12        0.00           0.00          ~     (all equal)
GetCacheColorFgHi         0.00           0.00          ~     (all equal)
GetCacheColorFgHi-4       0.00           0.00          ~     (all equal)
GetCacheColorFgHi-12      0.00           0.00          ~     (all equal)
GetCacheColorBg           0.00           0.00          ~     (all equal)
GetCacheColorBg-4         0.00           0.00          ~     (all equal)
GetCacheColorBg-12        0.00           0.00          ~     (all equal)
GetCacheColorBgHi         0.00           0.00          ~     (all equal)
GetCacheColorBgHi-4       0.00           0.00          ~     (all equal)
GetCacheColorBgHi-12      0.00           0.00          ~     (all equal)
ColorPrintFg              5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintFg-4            5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintFg-12           5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintFgHi            5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintFgHi-4          5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintFgHi-12         5.00 ± 0%      4.00 ± 0%  -20.00%  (p=0.000 n=10+10)
ColorPrintBgHi            6.00 ± 0%      4.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorPrintBgHi-4          6.00 ± 0%      4.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorPrintBgHi-12         6.00 ± 0%      4.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFg             9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFg-4           9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFg-12          9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFgHi           9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFgHi-4         9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringFgHi-12        9.00 ± 0%      6.00 ± 0%  -33.33%  (p=0.000 n=10+10)
ColorStringBgHi           10.0 ± 0%       6.0 ± 0%  -40.00%  (p=0.000 n=10+10)
ColorStringBgHi-4         10.0 ± 0%       6.0 ± 0%  -40.00%  (p=0.000 n=10+10)
ColorStringBgHi-12        10.0 ± 0%       6.0 ± 0%  -40.00%  (p=0.000 n=10+10)
```

## Todo

- [ ] Save/Return previous values
- [ ] Evaluate fmt.Formatter interface


## Credits

- [Fatih Arslan](https://github.com/fatih)
- Windows support via @mattn: [colorable](https://github.com/mattn/go-colorable)
- 2018- The color Authors.

## License

The MIT License (MIT) - see [`LICENSE.md`](https://github.com/zchee/color/blob/master/LICENSE.md) for more details


<!-- badge links -->
[circleci]: https://circleci.com/gh/zchee/workflows/color
[codecov]: https://codecov.io/gh/zchee/color
[godoc]: https://godoc.org/github.com/zchee/color
[tag]: https://github.com/zchee/color/releases
[ga]: https://github.com/zchee/color

[circleci-badge]: https://img.shields.io/circleci/build/github/zchee/color/master.svg?style=for-the-badge&label=CIRCLECI&logo=circleci?cacheSeconds=60
[godoc-badge]: https://img.shields.io/badge/godoc-reference-4F73B3.svg?style=for-the-badge&label=GODOC.ORG&logoWidth=25&logo=data%3Aimage%2Fsvg%2Bxml%3Bcharset%3Dutf-8%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MCIgaGVpZ2h0PSI0MCIgdmlld0JveD0iODUgNTUgMTIwIDEyMCI+PHBhdGggZmlsbD0iIzJEQkNBRiIgZD0iTTQwLjIgMTAxLjFjLS40IDAtLjUtLjItLjMtLjVsMi4xLTIuN2MuMi0uMy43LS41IDEuMS0uNWgzNS43Yy40IDAgLjUuMy4zLjZsLTEuNyAyLjZjLS4yLjMtLjcuNi0xIC42bC0zNi4yLS4xek0yNS4xIDExMC4zYy0uNCAwLS41LS4yLS4zLS41bDIuMS0yLjdjLjItLjMuNy0uNSAxLjEtLjVoNDUuNmMuNCAwIC42LjMuNS42bC0uOCAyLjRjLS4xLjQtLjUuNi0uOS42bC00Ny4zLjF6TTQ5LjMgMTE5LjVjLS40IDAtLjUtLjMtLjMtLjZsMS40LTIuNWMuMi0uMy42LS42IDEtLjZoMjBjLjQgMCAuNi4zLjYuN2wtLjIgMi40YzAgLjQtLjQuNy0uNy43bC0yMS44LS4xek0xNTMuMSA5OS4zYy02LjMgMS42LTEwLjYgMi44LTE2LjggNC40LTEuNS40LTEuNi41LTIuOS0xLTEuNS0xLjctMi42LTIuOC00LjctMy44LTYuMy0zLjEtMTIuNC0yLjItMTguMSAxLjUtNi44IDQuNC0xMC4zIDEwLjktMTAuMiAxOSAuMSA4IDUuNiAxNC42IDEzLjUgMTUuNyA2LjguOSAxMi41LTEuNSAxNy02LjYuOS0xLjEgMS43LTIuMyAyLjctMy43aC0xOS4zYy0yLjEgMC0yLjYtMS4zLTEuOS0zIDEuMy0zLjEgMy43LTguMyA1LjEtMTAuOS4zLS42IDEtMS42IDIuNS0xLjZoMzYuNGMtLjIgMi43LS4yIDUuNC0uNiA4LjEtMS4xIDcuMi0zLjggMTMuOC04LjIgMTkuNi03LjIgOS41LTE2LjYgMTUuNC0yOC41IDE3LTkuOCAxLjMtMTguOS0uNi0yNi45LTYuNi03LjQtNS42LTExLjYtMTMtMTIuNy0yMi4yLTEuMy0xMC45IDEuOS0yMC43IDguNS0yOS4zIDcuMS05LjMgMTYuNS0xNS4yIDI4LTE3LjMgOS40LTEuNyAxOC40LS42IDI2LjUgNC45IDUuMyAzLjUgOS4xIDguMyAxMS42IDE0LjEuNi45LjIgMS40LTEgMS43eiIvPjxwYXRoIGZpbGw9IiMyREJDQUYiIGQ9Ik0xODYuMiAxNTQuNmMtOS4xLS4yLTE3LjQtMi44LTI0LjQtOC44LTUuOS01LjEtOS42LTExLjYtMTAuOC0xOS4zLTEuOC0xMS4zIDEuMy0yMS4zIDguMS0zMC4yIDcuMy05LjYgMTYuMS0xNC42IDI4LTE2LjcgMTAuMi0xLjggMTkuOC0uOCAyOC41IDUuMSA3LjkgNS40IDEyLjggMTIuNyAxNC4xIDIyLjMgMS43IDEzLjUtMi4yIDI0LjUtMTEuNSAzMy45LTYuNiA2LjctMTQuNyAxMC45LTI0IDEyLjgtMi43LjUtNS40LjYtOCAuOXptMjMuOC00MC40Yy0uMS0xLjMtLjEtMi4zLS4zLTMuMy0xLjgtOS45LTEwLjktMTUuNS0yMC40LTEzLjMtOS4zIDIuMS0xNS4zIDgtMTcuNSAxNy40LTEuOCA3LjggMiAxNS43IDkuMiAxOC45IDUuNSAyLjQgMTEgMi4xIDE2LjMtLjYgNy45LTQuMSAxMi4yLTEwLjUgMTIuNy0xOS4xeiIvPjwvc3ZnPg==
[codecov-badge]: https://img.shields.io/codecov/c/github/zchee/color/master.svg?logo=codecov&style=for-the-badge&cacheSeconds=60
[release-badge]: https://img.shields.io/github/release/zchee/color.svg?logo=github&style=for-the-badge&cacheSeconds=60
[ga-badge]: https://gh-ga-beacon.appspot.com/UA-89201129-1/zchee/color?useReferer&pixel
