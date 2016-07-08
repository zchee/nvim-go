// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/garyburd/neovim-go/vim"
)

type pluginSpec struct {
	Type string            `msgpack:"type"`
	Name string            `msgpack:"name"`
	Sync bool              `msgpack:"sync"`
	Opts map[string]string `msgpack:"opts"`

	ServiceMethod string `msgpack:"-"`
	fn            interface{}
}

type handler struct {
	sm string
	fn interface{}
}

var (
	pluginSpecs = []*pluginSpec{}
	handlers    []*handler
)

func isSync(f interface{}) bool {
	t := reflect.TypeOf(f)
	return t.Kind() == reflect.Func && t.NumOut() > 0
}

func RegisterHandlers(v *vim.Vim, paths ...string) error {
	var mu sync.Mutex
	done := false
	err := v.RegisterHandler("specs", func(v *vim.Vim, path string) ([]*pluginSpec, error) {
		mu.Lock()
		defer mu.Unlock()
		if done {
			return []*pluginSpec{}, nil
		}
		done = true
		return pluginSpecs, nil
	})
	if err != nil {
		return err
	}
	for _, path := range paths {
		for _, s := range pluginSpecs {
			if err := v.RegisterHandler(path+s.ServiceMethod, s.fn); err != nil {
				return err
			}
		}
	}
	for _, h := range handlers {
		if err := v.RegisterHandler(h.sm, h.fn); err != nil {
			return err
		}
	}
	return nil
}

// Handle registers fn as a MessagePack RPC handler for the specified method
// name. The function signature for fn is one of
//
//  func(v *vim.Vim, {args}) ({resultType}, error)
//  func(v *vim.Vim, {args}) error
//  func(v *vim.Vim, {args})
//
// where {args} is zero or more arguments and {resultType} is the type of of a
// return value. Call the handler from Neovim using the rpcnotify and
// rpcrequest functions:
//
//  :help rpcrequest()
//  :help rpcnotify()
func Handle(method string, fn interface{}) {
	handlers = append(handlers, &handler{fn: fn, sm: method})
}

// FunctionOptions specifies function options.
type FunctionOptions struct {
	// Eval is an expression evaluated in Neovim. The result is passed the
	// handler function.
	Eval string
}

// HandleFunction registers fn as a handler for a Neovim function with the
// specified name. The name must be made of alphanumeric characters and '_',
// and must start with a capital letter. The function signature for fn is one
// of
//
//  func(v *vim.Vim, args {arrayType} [, eval {evalType}]) ({resultType}, error)
//  func(v *vim.Vim, args {arrayType} [, eval {evalType}]) error
//
// where {arrayType} is a type that can be unmarshaled from a MessagePack
// array, {evalType} is a type compatible with the Eval option expression and
// {resultType} is the type of function result.
//
// If options.Eval == "*", then HandleFunction constructs the expression to
// evaluate in Neovim from the type of fn's last argument. The last argument is
// assumed to be a pointer to a struct type with 'eval' field tags set to the
// expression to evaluate for each field. Nested structs are supported. The
// expression for the function
//
//  func example(v *vim.Vim, eval *struct{
//      GOPATH string `eval:"$GOPATH"`
//      Cwd    string `eval:"getcwd()"`
//  })
//
// is
//
//  {'GOPATH': $GOPATH, Cwd: getcwd()}
func HandleFunction(name string, options *FunctionOptions, fn interface{}) {
	m := make(map[string]string)
	if options != nil {
		if options.Eval != "" {
			m["eval"] = eval(options.Eval, fn)
		}
	}
	pluginSpecs = append(pluginSpecs, &pluginSpec{
		Type: "function",
		Name: name,
		Sync: isSync(fn),
		Opts: m,

		fn:            fn,
		ServiceMethod: ":function:" + name,
	})
}

// CommandOptions specifies command options.
type CommandOptions struct {

	// NArgs specifies the number command arguments.
	//
	//  0   No arguments are allowed
	//  1   Exactly one argument is required, it includes spaces
	//  *   Any number of arguments are allowed (0, 1, or many),
	//      separated by white space
	//  ?   0 or 1 arguments are allowed
	//  +   Arguments must be supplied, but any number are allowed
	NArgs string

	// Range specifies that the command accepts a range.
	//
	//  .   Range allowed, default is current line. The value
	//      "." is converted to "" for Neovim.
	//  %   Range allowed, default is whole file (1,$)
	//  N   A count (default N) which is specified in the line
	//      number position (like |:split|); allows for zero line
	//	    number.
	//
	//  :help :command-range
	Range string

	// Count specfies that thecommand accepts a count.
	//
	//  N   A count (default N) which is specified either in the line
	//	    number position, or as an initial argument (like |:Next|).
	//      Specifying -count (without a default) acts like -count=0
	//
	//  :help :command-count
	Count string

	// Addr sepcifies the domain for the range option
	//
	//  lines           Range of lines (this is the default)
	//  arguments       Range for arguments
	//  buffers         Range for buffers (also not loaded buffers)
	//  loaded_buffers  Range for loaded buffers
	//  windows         Range for windows
	//  tabs            Range for tab pages
	//
	//  :help command-addr
	Addr string

	// Bang specifies that the command can take a ! modifier (like :q or :w).
	Bang bool

	// Register specifes that the first argument to the command can be an
	// optional register name (like :del, :put, :yank).
	Register bool

	// Eval is evaluated in Neovim and the result is passed as an argument.
	Eval string

	// Bar specifies that the command can be followed by a "|" and another
	// command.  A "|" inside the command argument is not allowed then. Also
	// checks for a " to start a comment.
	Bar bool

	// Complete specifies command completion.
	//
	//  :help :command-complete
	Complete string
}

// HandleCommand registers fn as a handler for a Neovim command with the
// specified name. The name must be made of alphanumeric characters and '_',
// and must start with a capital letter.
///
// The arguments to fn function are:
//
//  v *vim.Vim
//  args []string       when options.NArgs != ""
//  range [2]int        when options.Range == "." or Range == "%"
//  range int           when options.Range == N or Count != ""
//  bang bool           when options.Bang == true
//  register string     when options.Register == true
//  eval interface{}    when options.Eval != ""
//
// The function fn must return an error.
//
// If options.Eval == "*", then HandleCommand constructs the expression to
// evaluate in Neovim from the type of fn's last argument. See the
// HandleFunction documentation for information on how the expression is
// generated.
func HandleCommand(name string, options *CommandOptions, fn interface{}) error {
	m := make(map[string]string)
	if options != nil {

		if options.NArgs != "" {
			m["nargs"] = options.NArgs
		}

		if options.Range != "" {
			if options.Range == "." {
				options.Range = ""
			}
			m["range"] = options.Range
		} else if options.Count != "" {
			m["count"] = options.Count
		}

		if options.Bang {
			m["bang"] = ""
		}

		if options.Register {
			m["register"] = ""
		}

		if options.Eval != "" {
			m["eval"] = eval(options.Eval, fn)
		}

		if options.Addr != "" {
			m["addr"] = options.Addr
		}

		if options.Bar {
			m["bar"] = ""
		}

		if options.Complete != "" {
			m["complete"] = options.Complete
		}
	}

	pluginSpecs = append(pluginSpecs, &pluginSpec{
		Type: "command",
		Name: name,
		Sync: isSync(fn),
		Opts: m,

		ServiceMethod: ":command:" + name,
		fn:            fn,
	})
	return nil
}

// AutocmdOptions specifies autocmd options.
type AutocmdOptions struct {
	// Group specifies the autocmd group.
	Group string

	// Pattern specifies an autocmd pattern.
	//
	//  :help autocmd-patterns
	Pattern string

	// Nested allows nested autocmds.
	//
	//  :help autocmd-nested
	Nested bool

	// Eval is evaluated in Neovim and the result is passed the the handler
	// function.
	Eval string
}

// HandleAutocmd registers fn as a handler for the specified autocmnd event.
//
// If options.Eval == "*", then HandleAutocmd constructs the expression to
// evaluate in Neovim from the type of fn's last argument. See the
// HandleFunction documentation for information on how the expression is
// generated.
func HandleAutocmd(event string, options *AutocmdOptions, fn interface{}) {
	pattern := ""
	m := make(map[string]string)
	if options != nil {

		if options.Group != "" {
			m["group"] = options.Group
		}

		if options.Pattern != "" {
			m["pattern"] = options.Pattern
			pattern = options.Pattern
		}

		if options.Nested {
			m["nested"] = "1"
		}

		if options.Eval != "" {
			m["eval"] = eval(options.Eval, fn)
		}

	}
	pluginSpecs = append(pluginSpecs, &pluginSpec{
		Type: "autocmd",
		Name: event,
		Sync: isSync(fn),
		Opts: m,

		fn:            fn,
		ServiceMethod: fmt.Sprintf(":autocmd:%s:%s", event, pattern),
	})

}

func eval(eval string, f interface{}) string {
	if eval != "*" {
		return eval
	}
	ft := reflect.TypeOf(f)
	if ft.Kind() != reflect.Func || ft.NumIn() < 1 {
		panic(`Eval: "*" option requires function with at least one argument`)
	}
	argt := ft.In(ft.NumIn() - 1)
	if argt.Kind() != reflect.Ptr || argt.Elem().Kind() != reflect.Struct {
		panic(`Eval: "*" option requires function with pointer to struct as last argument`)
	}
	return structEval(argt.Elem())
}

func structEval(t reflect.Type) string {
	buf := []byte{'{'}
	sep := ""
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Anonymous {
			panic(`Eval: "*" does not support anonymous fields`)
		}

		eval := sf.Tag.Get("eval")
		if eval == "" {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				eval = structEval(ft)
			}
		}

		if eval == "" {
			continue
		}

		name := strings.Split(sf.Tag.Get("msgpack"), ",")[0]
		if name == "" {
			name = sf.Name
		}

		buf = append(buf, sep...)
		buf = append(buf, "'"...)
		buf = append(buf, name...)
		buf = append(buf, "': "...)
		buf = append(buf, eval...)
		sep = ", "
	}
	buf = append(buf, '}')
	return string(buf)
}

type byServiceMethod []*pluginSpec

func (a byServiceMethod) Len() int           { return len(a) }
func (a byServiceMethod) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byServiceMethod) Less(i, j int) bool { return a[i].ServiceMethod < a[j].ServiceMethod }

func writePluginSpecs(w io.Writer) {
	// Sort for consistent order on output.
	sort.Sort(byServiceMethod(pluginSpecs))
	escape := strings.NewReplacer("'", "''").Replace

	fmt.Fprintf(w, "let s:specs = [\n")
	for _, spec := range pluginSpecs {
		sync := "0"
		if spec.Sync {
			sync = "1"
		}
		fmt.Fprintf(w, "\\ {'type': '%s', 'name': '%s', 'sync': %s, 'opts': {", spec.Type, spec.Name, sync)

		var keys []string
		for k := range spec.Opts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		optDelim := ""
		for _, k := range keys {
			fmt.Fprintf(w, "%s'%s': '%s'", optDelim, k, escape(spec.Opts[k]))
			optDelim = ", "
		}

		fmt.Fprintf(w, "}},\n")
	}
	fmt.Fprintf(w, "\\ ]\n")
}
