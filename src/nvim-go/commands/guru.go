// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// guru: a tool for answering questions about Go source code.
//
//    http://golang.org/s/oracle-design
//    http://golang.org/s/oracle-user-manual
//
// Run with -help flag or help subcommand for usage information.
//
package commands

import (
	"go/build"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"nvim-go/gb"
	"nvim-go/guru"
	"nvim-go/nvim"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
	"golang.org/x/tools/cmd/guru/serial"
)

func init() {
	plugin.HandleCommand("GoGuru", &plugin.CommandOptions{NArgs: "*", Complete: "customlist,GuruCompletelist", Eval: "[expand('%:p:h:h:h:h'), expand('%:p')]"}, Guru)
	plugin.HandleFunction("GuruCompletelist", &plugin.FunctionOptions{}, onComplete)
}

var (
	reflectFlag bool
)

type onGuruEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

func Guru(v *vim.Vim, args []string, eval *onGuruEval) error {
	defer gb.WithGoBuildForPath(eval.Cwd)()
	scopeFlag := []byte(eval.File)

	mode := args[0]
	pos, err := nvim.ByteOffset(v)
	if err != nil {
		return nvim.Echomsg(v, "%v", err)
	}

	gopath := os.Getenv("GOPATH")
	cmd := exec.Command("gb", "env", "GB_PROJECT_DIR")
	output, err := cmd.Output()
	if strings.Replace(string(output[:]), "\n", "", -1) != gopath {
		scopeFlag = output
	} else if err != nil {
		nvim.Echomsg(v, "could not find project root directory")
	}

	query := guru.Query{
		Mode:       mode,
		Pos:        eval.File + ":#" + strconv.FormatInt(int64(pos), 10),
		Build:      &build.Default,
		Scope:      strings.Split(string(scopeFlag[:]), ","),
		Reflection: reflectFlag,
	}

	if err := guru.Run(&query); err != nil {
		return nvim.Echomsg(v, "%s", err)
	}

	d := parseSerial(mode, query.Serial())
	if len(d) < 1 {
		return nvim.Echomsg(v, "0")
	}

	return nil
}

func parseSerial(mode string, s *serial.Result) map[string]interface{} {
	data := map[string]interface{}{
		"Mode": s.Mode,
	}

	switch mode {
	case "callees":
		data = map[string]interface{}{
			"Pos":             s.Callees.Pos,
			"Desc":            s.Callees.Desc,
			"Callees.Callees": s.Callees.Callees,
		}
	case "callers":
		data = map[string]interface{}{
			"Callers": s.Callers,
		}
	case "callstack":
		data = map[string]interface{}{
			"Pos":     s.Callstack.Pos,
			"Target":  s.Callstack.Target,
			"Callers": s.Callstack.Callers,
		}
	case "definition":
		data = map[string]interface{}{
			"ObjPos": s.Definition.ObjPos,
			"Desc":   s.Definition.Desc,
		}
	case "describe":
		data = map[string]interface{}{
			"Desc":         s.Describe.Desc,
			"Pos":          s.Describe.Pos,
			"Detail":       s.Describe.Detail,
			"Package":      s.Describe.Package,
			"Type":         s.Describe.Type,
			"Value.Type":   s.Describe.Value.Type,
			"Value.ObjPos": s.Describe.Value.ObjPos,
		}
	case "freevars":
		data = map[string]interface{}{
			"Freevars": s.Freevars,
		}
	case "implements":
		data = map[string]interface{}{
			"T":                       s.Implements.T,
			"AssignableTo":            s.Implements.AssignableTo,
			"AssignableFrom":          s.Implements.AssignableFrom,
			"AssignableFromPtr":       s.Implements.AssignableFromPtr,
			"Method":                  s.Implements.Method,
			"AssignableToMethod":      s.Implements.AssignableToMethod,
			"AssignableFromMethod":    s.Implements.AssignableFromMethod,
			"AssignableFromPtrMethod": s.Implements.AssignableFromPtrMethod,
		}
	case "peers":
		data = map[string]interface{}{
			"Pos":      s.Peers.Pos,
			"Type":     s.Peers.Type,
			"Allocs":   s.Peers.Allocs,
			"Sends":    s.Peers.Sends,
			"Receives": s.Peers.Receives,
			"Closes":   s.Peers.Closes,
		}
	case "pointsto":
		data = map[string]interface{}{
			"PointsTo": s.PointsTo,
		}
	case "referrers":
		data = map[string]interface{}{
			"Pos":    s.Referrers.Pos,
			"ObjPos": s.Referrers.ObjPos,
			"Desc":   s.Referrers.Desc,
			"Refs":   s.Referrers.Refs,
		}
	case "what":
		data = map[string]interface{}{
			"Enclosing":  s.What.Enclosing,
			"Modes":      s.What.Modes,
			"SrcDir":     s.What.SrcDir,
			"ImportPath": s.What.ImportPath,
		}
	case "whicherrs":
		data = map[string]interface{}{
			"ErrPos":    s.WhichErrs.ErrPos,
			"Globals":   s.WhichErrs.Globals,
			"Constants": s.WhichErrs.Constants,
			"Types":     s.WhichErrs.Types,
		}
	}

	return data
}

func onComplete(v *vim.Vim) ([]string, error) {
	return []string{
		"callers",
		"callees",
		"callstack",
		"definition",
		"describe",
		"freevars",
		"implements",
		"peers",
		"pointsto",
		"referrers",
		"what",
		"whicherrs",
	}, nil
}
