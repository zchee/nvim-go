// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vim implements a Neovim client.
//
// See the ./plugin package for additional functionality required for writing
// Neovim plugins.
package vim

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/garyburd/neovim-go/msgpack/rpc"
)

//go:generate go run genapi.go -out api.go

// Vim represents a remote instance of Neovim. It is safe to call *Vim methods
// concurrently.
type Vim struct {
	ep        *rpc.Endpoint
	mu        sync.Mutex
	channelID int

	// close is a hook for closing embedded Neovim process.
	close func() error
}

// Serve runs the MessagePack RPC server loop.
func (v *Vim) Serve() error {
	return v.ep.Serve()
}

// Close closes the client.
func (v *Vim) Close() error {
	err := v.ep.Close()
	if v.close != nil {
		errc := v.close()
		if err == nil {
			err = errc
		}
	}
	return err
}

// New create a Neovim client. When connecting to Neovim over stdio, use stdin
// as r and stdout as wc. When connecting to Neovim over a network connection,
// use the connection for both r and wc.
//
//  :help msgpack-rpc-connecting
func New(r io.Reader, wc io.WriteCloser, logf func(string, ...interface{})) (*Vim, error) {
	v := &Vim{}

	rwc := struct {
		io.Reader
		io.WriteCloser
	}{r, wc}

	var err error
	v.ep, err = rpc.NewEndpoint(rwc, withExtensions(), rpc.WithLogf(logf), rpc.WithFirstArg(v))
	if err != nil {
		return nil, err
	}

	return v, nil
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

// StartEmbeddedVim starts an embedded instance of Neovim using the specified
// arguments.
func StartEmbeddedVim(options *EmbedOptions) (*Vim, error) {
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

	v := &Vim{}
	rwc := struct {
		io.Reader
		io.WriteCloser
	}{outr, inw}

	v.ep, err = rpc.NewEndpoint(rwc, withExtensions(), rpc.WithLogf(options.Logf), rpc.WithFirstArg(v))
	if err != nil {
		return nil, err
	}
	closeOnExit = append(closeOnExit, v.ep)

	p, err := os.StartProcess(path,
		append([]string{path, "--embed"}, options.Args...),
		&os.ProcAttr{
			Env:   options.Env,
			Files: []*os.File{inr, outw},
		})
	if err != nil {
		return nil, err
	}

	closeOnExit = nil
	outw.Close()
	inr.Close()

	done := make(chan error, 1)
	go func() {
		done <- v.Serve()
		outr.Close()
		inw.Close()
	}()

	v.close = func() error {
		var errServe error
		select {
		case errServe = <-done:
		case <-time.After(5 * time.Second):
			p.Kill()
			errServe = errors.New("timeout waiting for nvim to exit")
		}
		state, err := p.Wait()
		if errServe != nil {
			return errServe
		}
		if err != nil {
			return err
		}
		if !state.Success() {
			return fmt.Errorf("%s", state)
		}
		return nil
	}

	return v, nil
}

// RegisterHandler registers fn as a MessagePack RPC handler for the named
// method. The function signature for fn is one of
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
//
// Plugin applications should use the Handler* methods in the ./plugin package
// to register handlers instead of this method.
func (v *Vim) RegisterHandler(method string, fn interface{}) error {
	return v.ep.RegisterHandler(method, fn)
}

// ChannelID returns Neovim's channel id for this client.
func (v *Vim) ChannelID() (int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.channelID != 0 {
		return v.channelID, nil
	}
	var info struct {
		ChannelID int `msgpack:",array"`
		Info      interface{}
	}
	if err := v.ep.Call("vim_get_api_info", &info); err != nil {
		return 0, err
	}
	v.channelID = info.ChannelID
	return v.channelID, nil
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
			el = append(el, fixError(c.ServiceMethod, c.Err))
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
