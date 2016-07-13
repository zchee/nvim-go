// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vim

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"time"

	"github.com/neovim-go/msgpack/rpc"
)

//go:generate go run genapi.go -out api.go

// Vim represents a remote instance of Neovim. It is safe to call *Vim methods
// concurrently.
type Vim struct {
	ep *rpc.Endpoint

	mu        sync.Mutex
	channelID int
}

// Serve serves incoming requests. Serve blocks until the Neovim disconnects or
// there is an error.
func (v *Vim) Serve() error {
	return v.ep.Serve()
}

// Close releases the resources used the client.
func (v *Vim) Close() error {
	return v.ep.Close()
}

// New creates a Neovim client. When connecting to Neovim over stdio, use stdin
// as r and stdout as w and c, When connecting to Neovim over a network connection,
// use the connection for r, w and c.
//
// The application must call Serve() to handle RPC requests and responses.
//
//  :help msgpack-rpc-connecting
func New(r io.Reader, w io.Writer, c io.Closer, logf func(string, ...interface{})) (*Vim, error) {
	ep, err := rpc.NewEndpoint(r, w, c, rpc.WithLogf(logf), withExtensions())
	if err != nil {
		return nil, err
	}
	return &Vim{ep: ep}, nil
}

// EmbedOptions specifies options for starting an embedded instance of Neovim.
type EmbedOptions struct {
	// Args specifies the command line arguments. Do not include the program
	// name (the first argument) or the --embed option.
	Args []string

	// Dir specifies the working directory of the command. The working
	// directory in the current process is sued if Dir is "".
	Dir string

	// Env specifies the environment of the Neovim process. The current process
	// environment is used if Env is nil.
	Env []string

	// Path is the path of the command to run. If Path = "", then
	// StartEmbeddedVim searches for "nvim" on $PATH.
	Path string

	Logf func(string, ...interface{})
}

type embedCloser struct {
	w io.WriteCloser
	p *os.Process
}

func (c *embedCloser) Close() error {
	err := c.w.Close()

	t := time.AfterFunc(5*time.Second, func() { c.p.Kill() })
	defer t.Stop()

	state, e := c.p.Wait()
	if e != nil {
		err = e
	}

	if err != nil || !state.Success() {
		err = fmt.Errorf("%s", state)
	}

	return err
}

// NewEmbedded starts an embedded instance of Neovim using the specified
// options.
func NewEmbedded(options *EmbedOptions) (*Vim, error) {
	var closeOnExit []io.Closer
	defer func() {
		for _, c := range closeOnExit {
			c.Close()
		}
	}()
	if options == nil {
		options = &EmbedOptions{}
	}

	path := options.Path
	if path == "" {
		var err error
		path, err = exec.LookPath("nvim")
		if err != nil {
			return nil, err
		}
	}

	outr, outw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	closeOnExit = append(closeOnExit, outr, outw)

	inr, inw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	closeOnExit = append(closeOnExit, inr, inw)

	c := &embedCloser{
		w: inw,
	}

	v, err := New(outr, inw, c, options.Logf)
	if err != nil {
		return nil, err
	}
	closeOnExit = append(closeOnExit, v.ep)

	c.p, err = os.StartProcess(path,
		append([]string{path, "--embed"}, options.Args...),
		&os.ProcAttr{
			Env:   options.Env,
			Files: []*os.File{inr, outw},
		})
	if err != nil {
		return nil, err
	}

	outw.Close()
	inr.Close()
	closeOnExit = nil
	return v, nil
}

// RegisterHandler registers fn as a MessagePack RPC handler for the named
// method. The function signature for fn is one of
//
//  func([v *vim.Vim,] {args}) ({resultType}, error)
//  func([v *vim.Vim,] {args}) error
//  func([v *vim.Vim,] {args})
//
// where {args} is zero or more arguments and {resultType} is the type of a
// return value. Call the handler from Neovim using the rpcnotify and
// rpcrequest functions:
//
//  :help rpcrequest()
//  :help rpcnotify()
//
// Plugin applications should use the Handler* methods in the ./plugin package
// to register handlers instead of this method.
func (v *Vim) RegisterHandler(method string, fn interface{}) error {
	var args []interface{}
	t := reflect.TypeOf(fn)
	if t.Kind() == reflect.Func && t.NumIn() > 0 && t.In(0) == reflect.TypeOf(v) {
		args = append(args, v)
	}
	return v.ep.Register(method, fn, args...)
}

// ChannelID returns Neovim's channel id for this client.
func (v *Vim) ChannelID() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.channelID != 0 {
		return v.channelID
	}
	var info struct {
		ChannelID int `msgpack:",array"`
		Info      interface{}
	}
	if err := v.ep.Call("vim_get_api_info", &info); err != nil {
		// TODO: log error and exit process?
	}
	v.channelID = info.ChannelID
	return v.channelID
}

func (v *Vim) call(sm string, result interface{}, args ...interface{}) error {
	return fixError(sm, v.ep.Call(sm, result, args...))
}

// NewPipeline creates a new pipeline.
func (v *Vim) NewPipeline() *Pipeline {
	return &Pipeline{ep: v.ep}
}

// Pipeline pipelines calls to Neovim. The underlying calls to Neovim execute
// and update result arguments asynchronous to the coller. Call the Wait method
// to wait for the calls to complete.
//
// Pipelines do not support concurrent calls by the application.
type Pipeline struct {
	ep    *rpc.Endpoint
	n     int
	done  chan *rpc.Call
	chans []chan *rpc.Call
}

const doneChunkSize = 32

func (p *Pipeline) call(sm string, result interface{}, args ...interface{}) {
	if p.n%doneChunkSize == 0 {
		done := make(chan *rpc.Call, doneChunkSize)
		p.done = done
		p.chans = append(p.chans, done)
	}
	p.n++
	p.ep.Go(sm, p.done, result, args...)
}

// Wait waits for all calls in the pipeline to complete. If there is more than
// one call in the pipeline, then Wait returns errors using type ErrorList.
func (p *Pipeline) Wait() error {
	var el ErrorList
	var done chan *rpc.Call
	useList := p.n > 1
	for i := 0; i < p.n; i++ {
		if i%doneChunkSize == 0 {
			done = p.chans[0]
			p.chans = p.chans[1:]
		}
		c := <-done
		if c.Err != nil {
			el = append(el, fixError(c.Method, c.Err))
		}
	}
	p.n = 0
	p.done = nil
	p.chans = nil
	switch {
	case len(el) == 0:
		return nil
	case useList:
		return el
	default:
		return el[0]
	}
}

func fixError(sm string, err error) error {
	if e, ok := err.(rpc.Error); ok {
		if a, ok := e.Value.([]interface{}); ok && len(a) == 2 {
			switch a[0] {
			case int64(exceptionError), uint64(exceptionError):
				return fmt.Errorf("nvim:%s exception: %v", sm, a[1])
			case int64(validationError), uint64(validationError):
				return fmt.Errorf("nvim:%s validation: %v", sm, a[1])
			}
		}
	}
	return err
}

// ErrorList is a list of errors.
type ErrorList []error

func (el ErrorList) Error() string {
	return el[0].Error()
}

// Call calls a vimscript function.
func (v *Vim) Call(fname string, result interface{}, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	return v.call("vim_call_function", result, fname, args)
}

// Call calls a vimscript function.
func (p *Pipeline) Call(fname string, result interface{}, args ...interface{}) {
	if args == nil {
		args = []interface{}{}
	}
	p.call("vim_call_function", result, fname, args)
}

// decodeExt decodes a MsgPack encoded number to an integer.
func decodeExt(p []byte) (int, error) {
	switch {
	case len(p) == 1 && p[0] <= 0x7f:
		return int(p[0]), nil
	case len(p) == 2 && p[0] == 0xcc:
		return int(p[1]), nil
	case len(p) == 3 && p[0] == 0xcd:
		return int(uint16(p[2]) | uint16(p[1])<<8), nil
	case len(p) == 5 && p[0] == 0xce:
		return int(uint32(p[4]) | uint32(p[3])<<8 | uint32(p[2])<<16 | uint32(p[1])<<24), nil
	case len(p) == 2 && p[0] == 0xd0:
		return int(int8(p[1])), nil
	case len(p) == 3 && p[0] == 0xd1:
		return int(int16(uint16(p[2]) | uint16(p[1])<<8)), nil
	case len(p) == 5 && p[0] == 0xd2:
		return int(int32(uint32(p[4]) | uint32(p[3])<<8 | uint32(p[2])<<16 | uint32(p[1])<<24)), nil
	case len(p) == 1 && p[0] >= 0xe0:
		return int(int8(p[0])), nil
	default:
		return 0, fmt.Errorf("nvimgo: error decoding extension bytes %x", p)
	}
}

// encodeExt encodes n to MsgPack format.
func encodeExt(n int) []byte {
	return []byte{0xd2, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}
