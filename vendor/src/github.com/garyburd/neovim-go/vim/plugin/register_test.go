// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.Handle("hello", func(v *vim.Vim, s string) (string, error) {
		return "Hello, " + s, nil
	})
}

func TestRegister(t *testing.T) {
	env := []string{}
	if v := os.Getenv("VIM"); v != "" {
		env = append(env, "VIM="+v)
	}
	v, err := vim.StartEmbeddedVim(&vim.EmbedOptions{
		Args: []string{"-u", "NONE", "-n"},
		Env:  env,
		Logf: t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()

	if err := plugin.RegisterHandlers(v, "x"); err != nil {
		t.Fatal(err)
	}

	cid, err := v.ChannelID()
	if err != nil {
		t.Fatal(err)
	}

	if err := v.Command(fmt.Sprintf(":call remote#host#RegisterPlugin('nvimgo', 'x', rpcrequest(%d, 'specs', 'x'))", cid)); err != nil {
		t.Error(err)
	}

	if err := v.Command(fmt.Sprintf(":call remote#host#Register('nvimgo', 'x', %d)", cid)); err != nil {
		t.Error(err)
	}

	{
		result, err := v.CommandOutput(":echo Hello('John', 'Doe')")
		if err != nil {
			t.Error(err)
		}
		expected := "\nHello, John Doe"
		if result != expected {
			t.Errorf("Hello returned %q, want %q", result, expected)
		}
	}

	{
		var result string
		if err := v.Call("rpcrequest", &result, cid, "hello", "world"); err != nil {
			t.Fatal(err)
		}

		expected := "Hello, world"
		if result != expected {
			t.Errorf("hello returned %q, want %q", result, expected)
		}
	}
}
