package main

import (
	"fmt"
	"log"

	"go/importer"
	"go/types"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	astPkg, err := importer.Default().Import("go/ast")
	if err != nil {
		log.Fatal(err)
	}

	astNodeInterface := astPkg.Scope().Lookup("Node").Type().Underlying().(*types.Interface)

	var tokenPkg *types.Package
	for _, pkg := range astPkg.Imports() {
		if pkg.Path() == "go/token" {
			tokenPkg = pkg
			break
		}
	}
	tokenPosType := tokenPkg.Scope().Lookup("Pos").Type()

	names := astPkg.Scope().Names()
	for _, name := range names {
		tn, ok := astPkg.Scope().Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}

		t, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}

		s, ok := t.Underlying().(*types.Struct)
		if !ok {
			continue
		}

		if !types.Implements(types.NewPointer(t), astNodeInterface) {
			continue
		}

		fmt.Printf("case *ast.%s:\n", t.Obj().Name())

		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)
			if types.Identical(f.Type(), tokenPosType) {
				fmt.Printf("normalize(&n.%s)\n", f.Name())
			} else if types.Implements(f.Type(), astNodeInterface) {
				fmt.Printf("if n.%s != nil { NormalizePos(n.%s) }\n", f.Name(), f.Name())
			} else if sl, ok := f.Type().(*types.Slice); ok && types.Implements(sl.Elem(), astNodeInterface) {
				fmt.Printf("for _, x := range n.%s { NormalizePos(x) }\n", f.Name())
			}
		}

		fmt.Println()
	}
}
