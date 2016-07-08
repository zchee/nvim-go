// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rpc implements MessagePack RPC.
package rpc

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/garyburd/neovim-go/msgpack"
)

const (
	requestMessage      = 0
	replyMessage        = 1
	notificationMessage = 2
)

var (
	errClosed   = errors.New("msgpack/rpc: session closed")
	errInternal = errors.New("msgpack/rpc: internal error")
)

type Error struct {
	Value interface{}
}

func (e Error) Error() string {
	return fmt.Sprintf("%v", e.Value)
}

type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Err           error
	Done          chan *Call
}

func (call *Call) done(e *Endpoint, err error) {
	call.Err = err
	select {
	case call.Done <- call:
		// ok
	default:
		e.logf("msgpack/rpc: done channel over capacity for method %s", call.ServiceMethod)
	}
}

type Endpoint struct {
	logf func(fmt string, args ...interface{})
	arg  reflect.Value

	dec *msgpack.Decoder

	packMu sync.Mutex
	enc    *msgpack.Encoder

	mu      sync.Mutex
	closed  bool
	id      uint64
	pending map[uint64]*Call
	closer  io.Closer

	handlersMu sync.RWMutex
	handlers   map[string]reflect.Value
}

func NewEndpoint(conn io.ReadWriteCloser, options ...Option) (*Endpoint, error) {
	e := &Endpoint{
		closer:   conn,
		enc:      msgpack.NewEncoder(conn),
		dec:      msgpack.NewDecoder(conn),
		pending:  make(map[uint64]*Call),
		handlers: make(map[string]reflect.Value),
		logf:     func(fmt string, args ...interface{}) {},
	}
	for _, option := range options {
		option.f(e)
	}
	return e, nil
}

type Option struct{ f func(*Endpoint) }

func WithExtensions(extensions msgpack.ExtensionMap) Option {
	return Option{func(e *Endpoint) {
		e.dec.SetExtensions(extensions)
	}}
}

func WithLogf(f func(fmt string, args ...interface{})) Option {
	return Option{func(e *Endpoint) {
		e.logf = f
	}}
}

func WithFirstArg(v interface{}) Option {
	return Option{func(e *Endpoint) {
		e.arg = reflect.ValueOf(v)
	}}
}

func (e *Endpoint) decodeUint(what string) (uint64, error) {
	if err := e.dec.Unpack(); err != nil {
		return 0, err
	}
	t := e.dec.Type()
	if t != msgpack.Uint && t != msgpack.Int {
		return 0, fmt.Errorf("msgpack/rpc: error decoding %s, found %s", what, e.dec.Type())
	}
	return e.dec.Uint(), nil
}

func (e *Endpoint) decodeString(what string) (string, error) {
	if err := e.dec.Unpack(); err != nil {
		return "", err
	}
	if e.dec.Type() != msgpack.String {
		return "", fmt.Errorf("msgpack/rpc: error decoding %s, found %s", what, e.dec.Type())
	}
	return e.dec.String(), nil
}

func (e *Endpoint) skip(n int) error {
	for i := 0; i < n; i++ {
		if err := e.dec.Unpack(); err != nil {
			return err
		}
		if err := e.dec.Skip(); err != nil {
			return err
		}
	}
	return nil
}

func (e *Endpoint) fatal(err error) error {
	e.Close()
	return err
}

func (e *Endpoint) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return errClosed
	}
	e.closed = true
	for _, call := range e.pending {
		call.done(e, errClosed)
	}
	e.pending = nil
	return e.closer.Close()
}

func (e *Endpoint) Call(serviceMethod string, reply interface{}, args ...interface{}) error {
	c := <-e.Go(serviceMethod, make(chan *Call, 1), reply, args...).Done
	return c.Err
}

func (e *Endpoint) Go(serviceMethod string, done chan *Call, reply interface{}, args ...interface{}) *Call {
	if args == nil {
		args = []interface{}{}
	}
	if done == nil {
		done = make(chan *Call, 1)
	} else if cap(done) == 0 {
		panic("unbuffered done channel")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}

	e.packMu.Lock()
	defer e.packMu.Unlock()

	e.mu.Lock()
	if e.closed {
		call.done(e, errClosed)
		e.mu.Unlock()
		return call
	}
	e.id = (e.id + 1) & 0x7fffffff
	id := e.id
	e.pending[id] = call
	e.mu.Unlock()

	if err := e.enc.Encode([]interface{}{requestMessage, id, serviceMethod, args}); err != nil {
		e.fatal(fmt.Errorf("msgpack/rpc: error encoding %s: %v", call.ServiceMethod, err))
	}

	return call
}

func (e *Endpoint) Notify(serviceMethod string, args ...interface{}) error {
	if args == nil {
		args = []interface{}{}
	}
	e.packMu.Lock()
	defer e.packMu.Unlock()
	if err := e.enc.Encode([]interface{}{notificationMessage, serviceMethod, args}); err != nil {
		e.fatal(fmt.Errorf("msgpack/rpc: error encoding %s: %v", serviceMethod, err))
	}
	return nil
}

var errorType = reflect.ValueOf(new(error)).Elem().Type()

func (e *Endpoint) RegisterHandler(serviceMethod string, function interface{}) error {
	v := reflect.ValueOf(function)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return errors.New("msgpack/rpc: handler not a function")
	}

	if e.arg.IsValid() && (t.NumIn() == 0 || t.In(0) != e.arg.Type()) {
		return fmt.Errorf("msgpack/rpc: first handler arg must be type %s", e.arg.Type())
	}

	if t.NumOut() > 2 || (t.NumOut() > 1 && t.Out(t.NumOut()-1) != errorType) {
		return errors.New("msgpack/rpc: handler return must be (), (error) or (valueType, error)")
	}
	e.handlersMu.Lock()
	e.handlers[serviceMethod] = v
	e.handlersMu.Unlock()
	return nil
}

func (e *Endpoint) reply(id uint64, replyErr error, reply interface{}) error {
	e.packMu.Lock()
	defer e.packMu.Unlock()

	err := e.enc.PackArrayLen(4)
	if err != nil {
		return err
	}

	err = e.enc.PackUint(replyMessage)
	if err != nil {
		return err
	}

	err = e.enc.PackUint(id)
	if err != nil {
		return err
	}

	if replyErr == nil {
		err = e.enc.PackNil()
	} else if ee, ok := replyErr.(Error); ok {
		err = e.enc.Encode(ee.Value)
	} else if ee, ok := replyErr.(msgpack.Marshaler); ok {
		err = ee.MarshalMsgPack(e.enc)
	} else {
		err = e.enc.PackString(replyErr.Error())
	}
	if err != nil {
		return err
	}

	return e.enc.Encode(reply)
}

func (e *Endpoint) Serve() error {
	for {
		if err := e.dec.Unpack(); err != nil {
			if err == io.EOF {
				err = nil
			}
			return e.fatal(err)
		}

		messageLen := e.dec.Len()
		if messageLen < 1 {
			return e.fatal(fmt.Errorf("msgpack/rpc: invalid message length %d", messageLen))
		}

		messageType, err := e.decodeUint("message type")
		if err != nil {
			return e.fatal(err)
		}

		switch messageType {
		case requestMessage:
			err = e.handleRequest(messageLen)
		case replyMessage:
			err = e.handleReply(messageLen)
		case notificationMessage:
			err = e.handleNotification(messageLen)
		default:
			err = fmt.Errorf("msgpack/rpc: unknown message type %d", messageType)
		}
		if err != nil {
			return e.fatal(err)
		}
	}
}

func (e *Endpoint) handleReply(messageLen int) error {
	if messageLen != 4 {
		// messageType, id, error, reply
		return fmt.Errorf("msgpack/rpc: invalid reply message length %d", messageLen)
	}

	id, err := e.decodeUint("response id")
	if err != nil {
		return err
	}

	e.mu.Lock()
	call := e.pending[id]
	delete(e.pending, id)
	e.mu.Unlock()

	if call == nil {
		e.logf("msgpack/rpc: no pending call for reply %d", id)
		return e.skip(2)
	}

	var errorValue interface{}
	if err := e.dec.Decode(&errorValue); err != nil {
		call.done(e, errInternal)
		return fmt.Errorf("msgpack/rpc: error decoding error value: %v", err)
	}

	if errorValue != nil {
		err := e.skip(1)
		call.done(e, Error{errorValue})
		return err
	}

	if call.Reply == nil {
		err = e.skip(1)
	} else {
		err = e.dec.Decode(call.Reply)
		if cvterr, ok := err.(*msgpack.DecodeConvertError); ok {
			call.done(e, cvterr)
			return nil
		}
	}

	if err != nil {
		call.done(e, errInternal)
		return fmt.Errorf("msgpack/rpc: error decoding reply: %v", err)
	}

	call.done(e, nil)
	return nil
}

func (e *Endpoint) createCall(f reflect.Value) (func([]reflect.Value) []reflect.Value, []reflect.Value, error) {
	t := f.Type()
	args := make([]reflect.Value, t.NumIn())
	argIndex := 0
	if e.arg.IsValid() {
		args[argIndex] = e.arg
		argIndex++
	}

	if err := e.dec.Unpack(); err != nil {
		return nil, nil, err
	}
	if e.dec.Type() != msgpack.ArrayLen {
		e.dec.Skip()
		return nil, nil, fmt.Errorf("msgpack/rpc: expected args array, found %s", e.dec.Type())
	}
	inputLen := e.dec.Len()

	var savedErr error

	// Decode plain arguments.

	n := t.NumIn()
	if t.IsVariadic() {
		n--
	}

	inputIndex := 0
	for argIndex < n {
		v := reflect.New(t.In(argIndex))
		args[argIndex] = v.Elem()
		argIndex++
		if inputIndex < inputLen {
			inputIndex++
			err := e.dec.Decode(v.Interface())
			if _, ok := err.(*msgpack.DecodeConvertError); ok {
				if savedErr == nil {
					savedErr = err
				}
			} else if err != nil {
				return nil, nil, err
			}
		}
	}

	if !t.IsVariadic() {

		// Skip extra arguments

		n := inputLen - inputIndex
		if n > 0 {
			err := e.skip(n)
			if err != nil {
				return nil, nil, err
			}
		}

		return f.Call, args, savedErr
	}

	if inputIndex >= inputLen {
		args[argIndex] = reflect.Zero(t.In(argIndex))
		return f.CallSlice, args, savedErr
	}

	n = inputLen - inputIndex
	v := reflect.MakeSlice(t.In(argIndex), n, n)
	args[argIndex] = v

	for i := 0; i < n; i++ {
		err := e.dec.Decode(v.Index(i).Addr().Interface())
		if _, ok := err.(*msgpack.DecodeConvertError); ok {
			if savedErr == nil {
				savedErr = err
			}
		} else if err != nil {
			return nil, nil, err
		}
	}

	return f.CallSlice, args, nil
}

func (e *Endpoint) handleRequest(messageLen int) error {
	if messageLen != 4 {
		// messageType, id, serviceMethod, args
		return fmt.Errorf("msgpack/rpc: invalid request message length %d", messageLen)
	}

	id, err := e.decodeUint("request id")
	if err != nil {
		return err
	}

	serviceMethod, err := e.decodeString("service method name")
	if err != nil {
		return err
	}

	e.handlersMu.RLock()
	f, ok := e.handlers[serviceMethod]
	e.handlersMu.RUnlock()

	if !ok {
		if err := e.skip(1); err != nil {
			return err
		}
		e.logf("msgpack/rpc: request service method %s not found", serviceMethod)
		return e.reply(id, fmt.Errorf("unknown request method: %s", serviceMethod), nil)
	}

	call, args, err := e.createCall(f)
	if _, ok := err.(*msgpack.DecodeConvertError); ok {
		e.logf("msgpack/rpc: %s: %v", serviceMethod, err)
		return e.reply(id, errors.New("invalid argument"), nil)
	} else if err != nil {
		return err
	}

	go func() {
		out := call(args)
		var replyErr error
		var replyVal interface{}
		switch f.Type().NumOut() {
		case 1:
			replyErr, _ = out[0].Interface().(error)
		case 2:
			replyVal = out[0].Interface()
			replyErr, _ = out[1].Interface().(error)
		}
		if err := e.reply(id, replyErr, replyVal); err != nil {
			e.fatal(err)
		}
	}()
	return nil
}

func (e *Endpoint) handleNotification(messageLen int) error {
	// messageType, serviceMethod, args
	if messageLen != 3 {
		return fmt.Errorf("msgpack/rpc: invalid notification message length %d", messageLen)
	}

	serviceMethod, err := e.decodeString("service method name")
	if err != nil {
		return err
	}

	e.handlersMu.RLock()
	f, ok := e.handlers[serviceMethod]
	e.handlersMu.RUnlock()

	if !ok {
		e.logf("msgpack/rpc: notification service method %s not found", serviceMethod)
		return e.skip(1)
	}

	call, args, err := e.createCall(f)
	if err != nil {
		return err
	}

	go func() {
		out := call(args)
		if len(out) > 0 {
			replyErr, _ := out[len(out)-1].Interface().(error)
			if replyErr != nil {
				e.logf("msgpack/rpc: service method %s returned %v", serviceMethod, replyErr)
			}
		}
	}()

	return nil
}
