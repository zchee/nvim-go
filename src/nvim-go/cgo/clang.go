package cgo

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"github.com/zchee/go-clang/clang"

	"nvim-go/nvim"
)

var (
	clangNomodifiable  = "go#clang#nomodifiable"
	vClangNomodifiable interface{}
	clangDebug         = "go#debug"
	vClangDebug        interface{}
)

func init() {
	plugin.HandleCommand("GoClang", &plugin.CommandOptions{Eval: "expand('%:p')"}, Goto)
}

// Goto jump to defined in cgo source code for current cursor
func Goto(v *vim.Vim, file string) error {
	useNomodifiable := false
	v.Var(clangNomodifiable, &vClangNomodifiable)
	if vClangNomodifiable.(int64) == int64(1) {
		useNomodifiable = true
	}
	// debug := false
	// v.Var(clangDebug, &vClangDebug)
	// if vClangDebug.(int64) == int64(1) {
	// debug = true
	// }

	index := clang.NewIndex(0, 1)
	defer index.Dispose()

	// fake := `#inclued <stdlib:h>\n\nint main() {\n\n}`

	tu := index.ParseTranslationUnit(file, []string{}, nil, 0)
	defer tu.Dispose()

	w, err := v.CurrentWindow()
	if err != nil {
		return nvim.Echoerr(v, err)
	}
	// buffre cursor
	bc, err := v.WindowCursor(w)
	if err != nil {
		return nvim.Echoerr(v, err)
	}

	cfile := tu.File(file)
	location := tu.Location(cfile, uint32(bc[0]), uint32(bc[1]))

	// TranslationUnit->Cursor
	cursor := tu.Cursor(location)

	// TranslationUnit->Cursor->Location
	referLocation := cursor.Referenced().Location()
	refFile, refLine, refColumn, _ := referLocation.ExpansionLocation()

	if refFile.Name() != "" {
		p := v.NewPipeline()
		var loclist []*nvim.ErrorlistData
		loclist = append(loclist, &nvim.ErrorlistData{
			FileName: refFile.Name(),
			LNum:     int(refLine),
			Col:      int(refColumn),
			Text:     refFile.Name(),
		})
		if err := nvim.SetLoclist(p, loclist); err != nil {
			nvim.Echomsg(v, "GoClang: %s", err)
		}

		v.Command("silent! ll! 1")
		if useNomodifiable {
			var result interface{}
			v.Option("nomodifiable", result)
		}
		v.FeedKeys("zz", "normal", false)
	} else {
		nvim.Echomsg(v, "GoClang: not found of Referenced Location")
	}

	return nil
}

/*
*
* go-clang-dump
* shows how to dump the AST of a C/C++ file via the Cursor visitor API
*
 */
var fname = flag.String("fname", "", "the file to analyze")

func Dump() {

	fmt.Printf(":: go-clang-dump...\n")
	flag.Parse()
	fmt.Printf(":: fname: %s\n", *fname)
	fmt.Printf(":: args: %v\n", flag.Args())

	if *fname == "" {
		flag.Usage()
		fmt.Printf("please provide a file name to analyze\n")

		os.Exit(1)
	}

	idx := clang.NewIndex(0, 1)
	defer idx.Dispose()

	args := []string{}
	if len(flag.Args()) > 0 && flag.Args()[0] == "-" {
		args = make([]string, len(flag.Args()[1:]))
		copy(args, flag.Args()[1:])
	}

	tu := idx.ParseTranslationUnit(*fname, args, nil, 0)
	defer tu.Dispose()

	fmt.Printf("tu: %s\n", tu.Spelling())
	cursor := tu.TranslationUnitCursor()
	fmt.Printf("cursor-isnull: %v\n", cursor.IsNull())
	fmt.Printf("cursor: %s\n", cursor.Spelling())
	fmt.Printf("cursor-kind: %s\n", cursor.Kind().Spelling())

	fmt.Printf("tu-fname: %s\n", tu.File(*fname).Name())

	cursor.Visit(func(cursor, parent clang.Cursor) clang.ChildVisitResult {
		if cursor.IsNull() {
			fmt.Printf("cursor: <none>\n")

			return clang.ChildVisit_Continue
		}

		fmt.Printf("%s: %s (%s)\n", cursor.Kind().Spelling(), cursor.Spelling(), cursor.USR())

		switch cursor.Kind() {
		case clang.Cursor_ClassDecl, clang.Cursor_EnumDecl, clang.Cursor_StructDecl, clang.Cursor_Namespace:
			return clang.ChildVisit_Recurse
		}

		return clang.ChildVisit_Continue
	})

	fmt.Printf(":: bye.\n")
}

/*
*
* go-clang-compdb
* dumps the content of a clang compilation database
*
 */

func CompileCommands() {
	if len(os.Args) <= 1 {
		fmt.Printf("**error: you need to give a directory containing a 'compile_commands.json' file\n")

		os.Exit(1)
	}

	dir := os.ExpandEnv(os.Args[1])
	fmt.Printf(":: inspecting [%s]...\n", dir)

	fname := filepath.Join(dir, "compile_commands.json")
	f, err := os.Open(fname)
	if err != nil {
		fmt.Printf("**error: could not open file [%s]: %v\n", fname, err)

		os.Exit(1)
	}
	f.Close()

	err, db := clang.FromDirectory(dir)
	if err != nil {
		fmt.Printf("**error: could not open compilation database at [%s]: %v\n", dir, err)

		os.Exit(1)
	}
	defer db.Dispose()

	cmds := db.AllCompileCommands()
	ncmds := cmds.Size()

	fmt.Printf(":: got %d compile commands\n", ncmds)

	for i := uint32(0); i < ncmds; i++ {
		cmd := cmds.Command(i)

		fmt.Printf("::  --- cmd=%d ---\n", i)
		fmt.Printf("::  dir= %q\n", cmd.Directory())

		nargs := cmd.NumArgs()
		fmt.Printf("::  nargs= %d\n", nargs)

		sargs := make([]string, 0, nargs)
		for iarg := uint32(0); iarg < nargs; iarg++ {
			arg := cmd.Arg(iarg)
			sfmt := "%q, "
			if iarg+1 == nargs {
				sfmt = "%q"
			}
			sargs = append(sargs, fmt.Sprintf(sfmt, arg))

		}

		fmt.Printf("::  args= {%s}\n", strings.Join(sargs, ""))
		if i+1 != ncmds {
			fmt.Printf("::\n")
		}
	}
	fmt.Printf(":: inspecting [%s]... [done]\n", dir)
}
