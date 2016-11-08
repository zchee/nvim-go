// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"log"
	"net"
	"os"
	"regexp"

	"nvim-go/config"

	"github.com/neovim/go-client/nvim"
)

// NewSocketClient creates the Neovim client over the socket session.
func NewSocketClient() *nvim.Nvim {
	var (
		v   *nvim.Nvim
		err error
	)

	addr := config.ServerName
	if addr == "" {
		return nil
	}

	v, err = dialNvim(addr)
	if err != nil {
		log.Println(err)
		return nil
	}

	return v
}

// NewEmbedClient creates the Neovim client over the embed api.
func NewEmbedClient(args []string, dir string, env []string) *nvim.Nvim {
	options := &nvim.EmbedOptions{
		Args: args,
		Dir:  dir,
		Env:  env,
	}

	v, err := nvim.NewEmbedded(options)
	if err != nil {
		log.Println(err)
		return nil
	}

	return v
}

// NewStdioClient creates the Neovim client over the stdio.
func NewStdioClient() *nvim.Nvim {
	v, err := nvim.New(os.Stdin, os.Stdout, os.Stdout, log.Printf)
	if err != nil {
		log.Println(err)
		return nil
	}

	go v.Serve()

	return v
}

var tcpAddrRe = regexp.MustCompile(`:\d+$`)

func dialNvim(addr string) (*nvim.Nvim, error) {
	network := "unix"
	if tcpAddrRe.MatchString(addr) {
		network = "tcp"
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	v, err := nvim.New(conn, conn, conn, log.Printf)
	if err != nil {
		return nil, err
	}
	go v.Serve()

	return v, nil
}
