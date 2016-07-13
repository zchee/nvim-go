// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program generates Neovim API methods in api.go.
//
// The program generates the code from data declared in this file instead of
// using the output from nvim --api-info. This approach allows names and types
// to be modified to create a more idiomatic and convenient API for Go
// programmers.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

var extensions = []*struct {
	Type string
	Code int
	Doc  string
}{
	{"Buffer", 0, `// Buffer represents a remote Neovim buffer.`},
	{"Window", 1, `// Window represents a remote Neovim window.`},
	{"Tabpage", 2, `// Tabpage represents a remote Neovim tabpage.`},
}

var varResultDoc = `
//
// If the result parameter is a pointer value, then the target of the pointer
// is set to the old value of the variable. The target of the pointer is set to
// nil if there was no previous value or the previous value was v:null.
`

type param struct{ Name, Type string }

var methods = []*struct {
	Name   string
	Sm     string
	Return string
	Doc    string
	Params []param
}{
	{
		Name:   "BufferLineCount",
		Sm:     "buffer_line_count",
		Return: "int",
		Params: []param{{"buffer", "Buffer"}},
		Doc:    `// BufferLineCount returns the number of lines in the buffer.`,
	},
	{
		Name:   "BufferLines",
		Sm:     "buffer_get_lines",
		Return: "[][]byte",
		Params: []param{{"buffer", "Buffer"}, {"start", "int"}, {"end", "int"}, {"strict", "bool"}},
		Doc: `
// BufferLines retrieves a line range from a buffer.
//
// Indexing is zero-based, end-exclusive. Negative indices are interpreted as
// length+1+index, i e -1 refers to the index past the end. So to get the last
// element set start=-2 and end=-1.
//
// Out-of-bounds indices are clamped to the nearest valid value, unless strict
// = true.
`,
	},
	{
		Name:   "SetBufferLines",
		Sm:     "buffer_set_lines",
		Params: []param{{"buffer", "Buffer"}, {"start", "int"}, {"end", "int"}, {"strict", "bool"}, {"replacement", "[][]byte"}},
		Doc: `
// SetBufferLines replaces a line range on a buffer.
//
// Indexing is zero-based, end-exclusive. Negative indices are interpreted as
// length+1+index, ie -1 refers to the index past the end. So to change or
// delete the last element set start=-2 and end=-1.
//
// To insert lines at a given index, set both start and end to the same index.
// To delete a range of lines, set replacement to an empty array.
//
// Out-of-bounds indices are clamped to the nearest valid value, unless strict
// = true.
`,
	},
	{
		Name:   "BufferVar",
		Sm:     "buffer_get_var",
		Return: "interface{}",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}},
		Doc:    `// BufferVar gets a buffer-scoped (b:) variable.`,
	},
	{
		Name:   "SetBufferVar",
		Sm:     "buffer_set_var",
		Return: "interface{}",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetBufferVar sets a buffer-scoped (b:) variable.` + varResultDoc,
	},
	{
		Name:   "DelBufferVar",
		Sm:     "buffer_del_var",
		Return: "interface{}",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// DelBufferVar removes a buffer-scoped (b:) variable.` + varResultDoc,
	},
	{
		Name:   "BufferOption",
		Sm:     "buffer_get_option",
		Return: "interface{}",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}},
		Doc:    `// BufferOption gets a buffer option value.`,
	},
	{
		Name:   "SetBufferOption",
		Sm:     "buffer_set_option",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}, {"value", "interface{}"}},
		Doc: `
// SetBufferOption sets a buffer option value. The value nil deletes the option
// in the case where there's a global fallback.
`,
	},
	{
		Name:   "BufferNumber",
		Sm:     "buffer_get_number",
		Return: "int",
		Params: []param{{"buffer", "Buffer"}},
		Doc:    `// BufferNumber gets a buffer's number.`,
	},
	{
		Name:   "BufferName",
		Sm:     "buffer_get_name",
		Return: "string",
		Params: []param{{"buffer", "Buffer"}},
		Doc:    `// BufferName gets the full file name of a buffer.`,
	},
	{
		Name:   "SetBufferName",
		Sm:     "buffer_set_name",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}},
		Doc: `
// SetBufferName sets the full file name of a buffer.
// BufFilePre/BufFilePost are triggered.
`,
	},
	{
		Name:   "IsBufferValid",
		Sm:     "buffer_is_valid",
		Return: "bool",
		Params: []param{{"buffer", "Buffer"}},
		Doc:    `// IsBufferValid returns true if the buffer is valid.`,
	},
	{
		Name:   "BufferMark",
		Sm:     "buffer_get_mark",
		Return: "[2]int",
		Params: []param{{"buffer", "Buffer"}, {"name", "string"}},
		Doc:    `// BufferMark returns the (row,col) of the named mark.`,
	},
	{
		Name:   "AddBufferHighlight",
		Sm:     "buffer_add_highlight",
		Return: "int",
		Params: []param{{"buffer", "Buffer"}, {"srcID", "int"}, {"hlGroup", "string"}, {"line", "int"}, {"startCol", "int"}, {"endCol", "int"}},
		Doc: `
// AddBufferHighlight adds a highlight to buffer and returns the source id of
// the highlight.
//
// AddBufferHighlight can be used for plugins which dynamically generate
// highlights to a buffer (like a semantic highlighter or linter). The function
// adds a single highlight to a buffer. Unlike matchaddpos() highlights follow
// changes to line numbering (as lines are inserted/removed above the
// highlighted line), like signs and marks do.
//
// The srcID is useful for batch deletion/updating of a set of highlights. When
// called with srcID = 0, an unique source id is generated and returned.
// Succesive calls can pass in it as srcID to add new highlights to the same
// source group. All highlights in the same group can then be cleared with
// ClearBufferHighlight. If the highlight never will be manually deleted pass
// in -1 for srcID.
//
// If hlGroup is the empty string no highlight is added, but a new srcID is
// still returned. This is useful for an external plugin to synchrounously
// request an unique srcID at initialization, and later asynchronously add and
// clear highlights in response to buffer changes.
//
// The startCol and endCol parameters specify the range of columns to
// highlight. Use endCol = -1 to highlight to the end of the line.
`,
	},
	{
		Name:   "ClearBufferHighlight",
		Sm:     "buffer_clear_highlight",
		Params: []param{{"buffer", "Buffer"}, {"srcID", "int"}, {"startLine", "int"}, {"endLine", "int"}},
		Doc: `
// ClearBufferHighlight clears highlights from a given source group and a range
// of lines.
//
// To clear a source group in the entire buffer, pass in 1 and -1 to startLine
// and endLine respectively.
//
// The lineStart and lineEnd parameters specify the range of lines to clear.
// The end of range is exclusive. Specify -1 to clear to the end of the file.
`,
	},
	{
		Name:   "TabpageWindows",
		Sm:     "tabpage_get_windows",
		Return: "[]Window",
		Params: []param{{"tabpage", "Tabpage"}},
		Doc:    `// TabpageWindows returns the windows in a tabpage.`,
	},
	{
		Name:   "TabpageVar",
		Sm:     "tabpage_get_var",
		Return: "interface{}",
		Params: []param{{"tabpage", "Tabpage"}, {"name", "string"}},
		Doc:    `// TabpageVar gets a tab-scoped (t:) variable.`,
	},
	{
		Name:   "SetTabpageVar",
		Sm:     "tabpage_set_var",
		Return: "interface{}",
		Params: []param{{"tabpage", "Tabpage"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetTabpageVar sets a tab-scoped (t:) variable.` + varResultDoc,
	},
	{
		Name:   "DelTabpageVar",
		Sm:     "tabpage_del_var",
		Return: "interface{}",
		Params: []param{{"tabpage", "Tabpage"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// DelTabpageVar removes a tab-scoped (t:) variable.` + varResultDoc,
	},
	{
		Name:   "TabpageWindow",
		Sm:     "tabpage_get_window",
		Return: "Window",
		Params: []param{{"tabpage", "Tabpage"}},
		Doc:    `// TabpageWindow gets the current window in a tab page.`,
	},
	{
		Name:   "IsTabpageValid",
		Sm:     "tabpage_is_valid",
		Return: "bool",
		Params: []param{{"tabpage", "Tabpage"}},
		Doc:    `// IsTabpageValid checks if a tab page is valid.`,
	},
	{
		Name:   "UIAttach",
		Sm:     "ui_attach",
		Params: []param{{"width", "int"}, {"height", "int"}, {"enableRGB", "bool"}},
		Doc: `
// UIAttach registers the client as a remote UI. After this method is called,
// the client will receive redraw notifications.
`,
	},
	{
		Name: "UIDetach",
		Sm:   "ui_detach",
		Doc:  `// UIDetach unregisters the client as a remote UI.`,
	},
	{
		Name:   "UITryResize",
		Sm:     "ui_try_resize",
		Params: []param{{"width", "int"}, {"height", "int"}},
		Doc: `
// UITryResize notifies Neovim that the client window has resized. If possible,
// Neovim will send a redraw request to resize.
`,
	},
	{
		Name:   "Command",
		Sm:     "vim_command",
		Params: []param{{"str", "string"}},
		Doc:    `// Command executes a single ex command.`,
	},
	{
		Name:   "FeedKeys",
		Sm:     "vim_feedkeys",
		Params: []param{{"keys", "string"}, {"mode", "string"}, {"escapeCsi", "bool"}},
		Doc: `
// FeedKeys Pushes keys to the Neovim user input buffer. Options can be a string
// with the following character flags:
//
//  m:  Remap keys. This is default.
//  n:  Do not remap keys.
//  t:  Handle keys as if typed; otherwise they are handled as if coming from a
//     mapping. This matters for undo, opening folds, etc. 
`,
	},
	{
		Name:   "Input",
		Sm:     "vim_input",
		Return: "int",
		Params: []param{{"keys", "string"}},
		Doc: `
// Input pushes bytes to the Neovim low level input buffer.
// 
// Unlike FeedKeys, this uses the lowest level input buffer and the call is not
// deferred. It returns the number of bytes actually written(which can be less
// than what was requested if the buffer is full).
`,
	},
	{
		Name:   "ReplaceTermcodes",
		Sm:     "vim_replace_termcodes",
		Return: "string",
		Params: []param{{"str", "string"}, {"fromPart", "bool"}, {"doLt", "bool"}, {"special", "bool"}},
		Doc: `
// ReplaceTermcodes replaces any terminal code strings by byte sequences. The
// returned sequences are Neovim's internal representation of keys, for example:
//
//  <esc> -> '\x1b'
//  <cr>  -> '\r'
//  <c-l> -> '\x0c'
//  <up>  -> '\x80ku'
//
// The returned sequences can be used as input to feedkeys.
`,
	},
	{
		Name:   "CommandOutput",
		Sm:     "vim_command_output",
		Return: "string",
		Params: []param{{"str", "string"}},
		Doc: `
// CommandOutput executes a single ex command and returns the output.
`,
	},
	{
		Name:   "Eval",
		Sm:     "vim_eval",
		Return: "interface{}",
		Params: []param{{"str", "string"}},
		Doc: `
// Eval evaluates the expression str using the Vim internal expression
// evaluator. 
//
//  :help expression
`,
	},
	{
		Name:   "Strwidth",
		Sm:     "vim_strwidth",
		Return: "int",
		Params: []param{{"str", "string"}},
		Doc: `
// Strwidth returns the number of display cells the string occupies. Tab is
// counted as one cell.
`,
	},
	{
		Name:   "RuntimePaths",
		Sm:     "vim_list_runtime_paths",
		Return: "[]string",
		Doc: `
// RuntimePaths returns a list of paths contained in the runtimepath option.
`,
	},
	{
		Name:   "ChangeDirectory",
		Sm:     "vim_change_directory",
		Params: []param{{"dir", "string"}},
		Doc:    `// ChangeDirectory changes Vim working directory.`,
	},
	{
		Name:   "CurrentLine",
		Sm:     "vim_get_current_line",
		Return: "[]byte",
		Doc:    `// CurrentLine gets the current line in the current buffer.`,
	},
	{
		Name:   "SetCurrentLine",
		Sm:     "vim_set_current_line",
		Params: []param{{"line", "[]byte"}},
		Doc:    `// SetCurrentLine sets the current line in the current buffer.`,
	},
	{
		Name: "DeleteCurrentLine",
		Sm:   "vim_del_current_line",
		Doc:  `// DeleteCurrentLine deletes the current line in the current buffer.`,
	},
	{
		Name:   "Var",
		Sm:     "vim_get_var",
		Return: "interface{}",
		Params: []param{{"name", "string"}},
		Doc:    `// Var gets a global (g:) variable.`,
	},
	{
		Name:   "SetVar",
		Sm:     "vim_set_var",
		Return: "interface{}",
		Params: []param{{"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetVar sets a global (g:) variable.` + varResultDoc,
	},
	{
		Name:   "DelVar",
		Sm:     "vim_del_var",
		Return: "interface{}",
		Params: []param{{"name", "string"}, {"value", "interface{}"}},
		Doc:    `// DelVar removes a global (g:) variable.` + varResultDoc,
	},
	{
		Name:   "Vvar",
		Sm:     "vim_get_vvar",
		Return: "interface{}",
		Params: []param{{"name", "string"}},
		Doc:    `// Vvar gets a vim (v:) variable.`,
	},
	{
		Name:   "Option",
		Sm:     "vim_get_option",
		Return: "interface{}",
		Params: []param{{"name", "string"}},
		Doc:    `// Option gets an option.`,
	},
	{
		Name:   "SetOption",
		Sm:     "vim_set_option",
		Params: []param{{"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetOption sets an option.`,
	},
	{
		Name:   "WriteOut",
		Sm:     "vim_out_write",
		Params: []param{{"str", "string"}},
		Doc: `
// WriteOut writes a message to vim output buffer. The string is split and
// flushed after each newline. Incomplete lines are kept for writing later.
`,
	},
	{
		Name:   "WriteErr",
		Sm:     "vim_err_write",
		Params: []param{{"str", "string"}},
		Doc: `
// WriteErr writes a message to vim error buffer. The string is split and
// flushed after each newline. Incomplete lines are kept for writing later.
`,
	},
	{
		Name:   "ReportError",
		Sm:     "vim_report_error",
		Params: []param{{"str", "string"}},
		Doc:    `// ReportError writes prints str and a newline as an error message.`,
	},
	{
		Name:   "Buffers",
		Sm:     "vim_get_buffers",
		Return: "[]Buffer",
		Doc:    `// Buffers returns the current list of buffers.`,
	},
	{
		Name:   "CurrentBuffer",
		Sm:     "vim_get_current_buffer",
		Return: "Buffer",
		Doc:    `// CurrentBuffer returns the current buffer.`,
	},
	{
		Name:   "SetCurrentBuffer",
		Sm:     "vim_set_current_buffer",
		Params: []param{{"buffer", "Buffer"}},
		Doc:    `// SetCurrentBuffer sets the current buffer.`,
	},
	{
		Name:   "Windows",
		Sm:     "vim_get_windows",
		Return: "[]Window",
		Doc:    `// Windows returns the current list of windows.`,
	},
	{
		Name:   "CurrentWindow",
		Sm:     "vim_get_current_window",
		Return: "Window",
		Doc:    `// CurrentWindow returns the current window.`,
	},
	{
		Name:   "SetCurrentWindow",
		Sm:     "vim_set_current_window",
		Params: []param{{"window", "Window"}},
		Doc:    `// SetCurrentWindow sets the current window.`,
	},
	{
		Name:   "Tabpages",
		Sm:     "vim_get_tabpages",
		Return: "[]Tabpage",
		Doc:    `// Tabpages returns the current list of tabpages.`,
	},
	{
		Name:   "CurrentTabpage",
		Sm:     "vim_get_current_tabpage",
		Return: "Tabpage",
		Doc:    `// CurrentTabpage returns the current tabpage.`,
	},
	{
		Name:   "SetCurrentTabpage",
		Sm:     "vim_set_current_tabpage",
		Params: []param{{"tabpage", "Tabpage"}},
		Doc:    `// SetCurrentTabpage sets the current tabpage.`,
	},
	{
		Name:   "Subscribe",
		Sm:     "vim_subscribe",
		Params: []param{{"event", "string"}},
		Doc:    `// Subscribe subscribes to a Neovim event.`,
	},
	{
		Name:   "Unsubscribe",
		Sm:     "vim_unsubscribe",
		Params: []param{{"event", "string"}},
		Doc:    `// Unsubscribe unsubscribes to a Neovim event.`,
	},
	{
		Name:   "NameToColor",
		Sm:     "vim_name_to_color",
		Return: "int",
		Params: []param{{"name", "string"}},
	},
	{
		Name:   "ColorMap",
		Sm:     "vim_get_color_map",
		Return: "map[string]interface{}",
	},
	{
		Name:   "APIInfo",
		Sm:     "vim_get_api_info",
		Return: "[]interface{}",
	},
	{
		Name:   "WindowBuffer",
		Sm:     "window_get_buffer",
		Return: "Buffer",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowBuffer returns the current buffer in a window.`,
	},
	{
		Name:   "WindowCursor",
		Sm:     "window_get_cursor",
		Return: "[2]int",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowCursor returns the cursor position in the window.`,
	},
	{
		Name:   "SetWindowCursor",
		Sm:     "window_set_cursor",
		Params: []param{{"window", "Window"}, {"pos", "[2]int"}},
		Doc:    `// SetWindowCursor sets the cursor position in the window to the given position.`,
	},
	{
		Name:   "WindowHeight",
		Sm:     "window_get_height",
		Return: "int",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowHeight returns the window height.`,
	},
	{
		Name:   "SetWindowHeight",
		Sm:     "window_set_height",
		Params: []param{{"window", "Window"}, {"height", "int"}},
		Doc:    `// SetWindowHeight sets the window height.`,
	},
	{
		Name:   "WindowWidth",
		Sm:     "window_get_width",
		Return: "int",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowWidth returns the window width.`,
	},
	{
		Name:   "SetWindowWidth",
		Sm:     "window_set_width",
		Params: []param{{"window", "Window"}, {"width", "int"}},
		Doc:    `// SetWindowWidth sets the window width.`,
	},
	{
		Name:   "WindowVar",
		Sm:     "window_get_var",
		Return: "interface{}",
		Params: []param{{"window", "Window"}, {"name", "string"}},
		Doc:    `// WindowVar gets a window-scoped (w:) variable.`,
	},
	{
		Name:   "SetWindowVar",
		Sm:     "window_set_var",
		Return: "interface{}",
		Params: []param{{"window", "Window"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetWindowVar sets a window-scoped (w:) variable.` + varResultDoc,
	},
	{
		Name:   "DelWindowVar",
		Sm:     "window_del_var",
		Return: "interface{}",
		Params: []param{{"window", "Window"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// DelWindowVar removes a window-scoped (w:) variable.` + varResultDoc,
	},
	{
		Name:   "WindowOption",
		Sm:     "window_get_option",
		Return: "interface{}",
		Params: []param{{"window", "Window"}, {"name", "string"}},
		Doc:    `// WindowOption gets a window option.`,
	},
	{
		Name:   "SetWindowOption",
		Sm:     "window_set_option",
		Params: []param{{"window", "Window"}, {"name", "string"}, {"value", "interface{}"}},
		Doc:    `// SetWindowOption sets a window option.`,
	},
	{
		Name:   "WindowPosition",
		Sm:     "window_get_position",
		Return: "[2]int",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowPosition gets the window position in display cells. First position is zero.`,
	},
	{
		Name:   "WindowTabpage",
		Sm:     "window_get_tabpage",
		Return: "Tabpage",
		Params: []param{{"window", "Window"}},
		Doc:    `// WindowTabpage gets the tab page that contains the window.`,
	},
	{
		Name:   "IsWindowValid",
		Sm:     "window_is_valid",
		Return: "bool",
		Params: []param{{"window", "Window"}},
		Doc:    `// IsWindowValid returns true if the window is valid.`,
	},
}

var templ = template.Must(template.New("").Parse(`// Code generated by 'go generate'

package vim

import (
    "fmt"
    "reflect"

    "github.com/neovim-go/msgpack"
    "github.com/neovim-go/msgpack/rpc"
)

const (
    exceptionError  = 0
    validationError = 1
)

func withExtensions() rpc.Option {
	return rpc.WithExtensions(msgpack.ExtensionMap{
{{range .Extensions}}
		{{.Code}}: func(p []byte) (interface{}, error) {
			x, err := decodeExt(p)
			return {{.Type}}(x), err
		},
{{end}}
	})
}

{{range .Extensions}}
{{.Doc}}
type {{.Type}} int

func (x *{{.Type}}) UnmarshalMsgPack(dec *msgpack.Decoder) error {
	if dec.Type() != msgpack.Extension || dec.Extension() != {{.Code}} {
		err := &msgpack.DecodeConvertError{
			SrcType:  dec.Type(),
			DestType: reflect.TypeOf(x),
		}
		dec.Skip()
		return err
	}
	n, err := decodeExt(dec.BytesNoCopy())
	*x = {{.Type}}(n)
	return err
}

func (x {{.Type}}) MarshalMsgPack(enc *msgpack.Encoder) error {
	return enc.PackExtension({{.Code}}, encodeExt(int(x)))
}

func (x {{.Type}}) String() string {
	return fmt.Sprintf("{{.Type}}:%d", int(x))
}
{{end}}

{{range .Methods}}
{{if eq "interface{}" .Return}}
{{.Doc}}
func (v *Vim) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}} result interface{}) error {
    return v.call("{{.Sm}}", result, {{range .Params}}{{.Name}},{{end}})
}

{{.Doc}}
func (p *Pipeline) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}} result interface{}) {
    p.call("{{.Sm}}", result, {{range .Params}}{{.Name}},{{end}})
}
{{else if .Return}}
{{.Doc}}
func (v *Vim) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}}) ({{.Return}}, error) {
    var result {{.Return}}
    err := v.call("{{.Sm}}", {{if .Return}}&result{{else}}nil{{end}}, {{range .Params}}{{.Name}},{{end}})
    return result, err
}
{{.Doc}}
func (p *Pipeline) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}} result *{{.Return}}) {
    p.call("{{.Sm}}", result, {{range .Params}}{{.Name}},{{end}})
}
{{else}}
{{.Doc}}
func (v *Vim) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}}) error {
    return v.call("{{.Sm}}", nil, {{range .Params}}{{.Name}},{{end}})
}
{{.Doc}}
func (p *Pipeline) {{.Name}}({{range .Params}}{{.Name}} {{.Type}},{{end}}) {
    p.call("{{.Sm}}", nil, {{range .Params}}{{.Name}},{{end}})
}
{{end}}
{{end}}
`))

func checkMethods() {
	expectedFirstParam := map[string]param{
		"buffer":  param{"buffer", "Buffer"},
		"tabpage": param{"tabpage", "Tabpage"},
		"window":  param{"window", "Window"},
	}
	serviceMethods := make(map[string]bool)

	for _, m := range methods {

		if serviceMethods[m.Sm] {
			log.Fatalf("Duplicate service method %s", m.Sm)
		}
		serviceMethods[m.Sm] = true

		if p := expectedFirstParam[strings.Split(m.Sm, "_")[0]]; p.Name != "" {
			if len(m.Params) < 1 || m.Params[0] != p {
				log.Fatalf("Service method %s does not have param[0] = %v", m.Sm, p)
			}
		}
	}
}

func main() {
	log.SetFlags(0)
	outFile := flag.String("out", "", "Output file")
	flag.Parse()

	for _, m := range methods {
		m.Doc = strings.TrimSpace(m.Doc)
	}

	checkMethods()

	var buf bytes.Buffer
	if err := templ.Execute(&buf, map[string]interface{}{
		"Methods":    methods,
		"Extensions": extensions,
	}); err != nil {
		log.Fatalf("error executing template: %v", err)
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		for i, p := range bytes.Split(buf.Bytes(), []byte("\n")) {
			fmt.Fprintf(os.Stderr, "%d: %s\n", i+1, p)
		}
		log.Fatalf("error formating source: %v", err)
	}

	f := os.Stdout
	if *outFile != "" {
		f, err = os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	f.Write(out)
}
