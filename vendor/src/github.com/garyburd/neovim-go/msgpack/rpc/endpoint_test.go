// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
)

func clientServer(t *testing.T, options ...Option) (*Endpoint, *Endpoint, func()) {
	var wg sync.WaitGroup
	wg.Add(2)

	options = append(options, WithLogf(t.Logf))

	serverConn, clientConn := net.Pipe()

	server, err := NewEndpoint(serverConn, options...)
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewEndpoint(clientConn, options...)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil && err != io.ErrClosedPipe {
			t.Logf("server: %v", err)
		}
		wg.Done()
	}()

	go func() {
		err := client.Serve()
		if err != nil && err != io.ErrClosedPipe {
			t.Logf("client: %v", err)
		}
		wg.Done()
	}()

	cleanup := func() {
		server.Close()
		client.Close()
		wg.Wait()
	}

	return client, server, cleanup
}

func TestEndpoint(t *testing.T) {
	client, server, cleanup := clientServer(t)
	defer cleanup()

	if err := server.RegisterHandler("add", func(a, b int) (int, error) { return a + b, nil }); err != nil {
		t.Fatal(err)
	}

	// Call.

	var sum int
	if err := client.Call("add", &sum, 1, 2); err != nil {
		t.Fatal(err)
	}

	if sum != 3 {
		t.Errorf("sum = %d, want %d", sum, 3)
	}

	// Notification.

	notifCh := make(chan string, 1)
	if err := server.RegisterHandler("n1", func(s string) { notifCh <- s }); err != nil {
		t.Fatal(err)
	}

	if err := client.Notify("n1", "hello"); err != nil {
		t.Fatal(err)
	}

	s := <-notifCh
	if s != "hello" {
		t.Fatal("no nello")
	}
}

var argsTests = []struct {
	sm     string
	args   []interface{}
	result []string
}{
	{"n", []interface{}{}, []string{"", ""}},
	{"n", []interface{}{"a"}, []string{"a", ""}},
	{"n", []interface{}{"a", "b"}, []string{"a", "b"}},
	{"n", []interface{}{"a", "b", "c"}, []string{"a", "b"}},

	{"v", []interface{}{}, []string{"", ""}},
	{"v", []interface{}{"a"}, []string{"a", ""}},
	{"v", []interface{}{"a", "b"}, []string{"a", "b"}},
	{"v", []interface{}{"a", "b", "x1"}, []string{"a", "b", "x1"}},
	{"v", []interface{}{"a", "b", "x1", "x2"}, []string{"a", "b", "x1", "x2"}},
	{"v", []interface{}{"a", "b", "x1", "x2", "x3"}, []string{"a", "b", "x1", "x2", "x3"}},

	{"a", []interface{}{}, []string(nil)},
	{"a", []interface{}{"x1", "x2", "x3"}, []string{"x1", "x2", "x3"}},
}

func TestArgs(t *testing.T) {
	client, server, cleanup := clientServer(t)
	defer cleanup()

	if err := server.RegisterHandler("n", func(a, b string) ([]string, error) {
		return append([]string{a, b}), nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := server.RegisterHandler("v", func(a, b string, x ...string) ([]string, error) {
		return append([]string{a, b}, x...), nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := server.RegisterHandler("a", func(x ...string) ([]string, error) {
		return x, nil
	}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range argsTests {
		var result []string
		if err := client.Call(tt.sm, &result, tt.args...); err != nil {
			t.Errorf("%s(%v) returned error %v", tt.sm, tt.args, err)
			continue
		}

		if !reflect.DeepEqual(result, tt.result) {
			t.Errorf("%s(%v) returned %#v, want %#v", tt.sm, tt.args, result, tt.result)
		}
	}
}

func TestFirstArg(t *testing.T) {
	client, server, cleanup := clientServer(t, WithFirstArg("hello"))
	defer cleanup()

	err := server.RegisterHandler("f", func(hello string) error {
		if hello != "hello" {
			t.Fatal("first arg not equal to 'hello'")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := client.Call("f", nil); err != nil {
		t.Fatal(err)
	}
}
