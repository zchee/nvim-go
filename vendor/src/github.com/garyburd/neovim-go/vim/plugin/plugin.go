// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package plugin is a Neovim remote plugin host.
//
// A plugin application registers one or more handlers for Neovim functions,
// commands and autocommands using the Handle* functions in this package.
// After registering the handlers, the application calls the Main function to
// run the plugin host.
//
// Use the default logger in the standard log package for logging in plugin
// applications. If the environment variable NEOVIM_GO_LOG_FILE is set, then
// the default logger is configured to append to the file specified by the
// environment variable.
package plugin

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/garyburd/neovim-go/vim"
)

var doWritePluginSpecs = flag.Bool("specs", false, "Write plugin specs to stdout")

// Main implements the main function for a Neovim remote plugin. The function
// creates a Neovim peer, registers RPC handlers and starts the peer server
// loop.
func Main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *doWritePluginSpecs {
		writePluginSpecs(os.Stdout)
		return
	}

	stdout := os.Stdout
	if fname := os.Getenv("NEOVIM_GO_LOG_FILE"); fname != "" {
		f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		os.Stdout = f
		os.Stderr = f
		log.SetOutput(f)
		log.SetPrefix(fmt.Sprintf("%8d ", os.Getpid()))
		log.Print("Plugin Start")
		defer log.Print("Plugin Exit")
	} else {
		log.SetFlags(0)
		os.Stdout = os.Stderr
	}

	v, err := vim.New(os.Stdin, stdout, log.Printf)
	if err != nil {
		log.Fatal(err)
	}

	var paths []string
	if len(os.Args) > 1 {
		paths = os.Args[1:]
	}

	if err := RegisterHandlers(v, paths...); err != nil {
		log.Fatal(err)
	}

	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
}
