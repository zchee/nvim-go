package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go/format"
	"go/parser"
	"go/printer"
	"golang.org/x/tools/go/loader"

	"github.com/motemen/go-iferr"
)

func main() {
	write := flag.Bool("w", false, "rewrite input files in place")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: goiferr [-w] <args>...")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, loader.FromArgsUsage)
	}
	flag.Parse()

	conf := loader.Config{
		AllowErrors: true,
		ParserMode:  parser.ParseComments,
	}
	conf.TypeChecker.Error = func(err error) {
		log.Printf("error (ignored): %s", err)
	}
	conf.FromArgs(flag.Args(), true)

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		for _, f := range pkg.Files {
			filename := prog.Fset.File(f.Pos()).Name()
			fmt.Fprintf(os.Stderr, "=== %s\n", filename)
			iferr.RewriteFile(prog.Fset, f, pkg.Info)
			if *write {
				fh, err := os.Create(filename)
				if err != nil {
					log.Fatal(err)
				}

				format.Node(fh, prog.Fset, f)
			} else {
				printer.Fprint(os.Stdout, prog.Fset, f)
			}
		}
	}
}
