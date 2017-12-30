package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

const hello = `package main

import "fmt"

func main() {
        fmt.Println("Hello, world")
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fake.go", hello, parser.Mode(0))
	if err != nil {
		log.Println(err)
	}

	ast.Print(fset, f)
}
