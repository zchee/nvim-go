// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nvim-go is a msgpack remote plugin for Neovim
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // For pprof debugging.
	"os"
	"runtime"

	"nvim-go/autocmd"
	"nvim-go/command"
	"nvim-go/command/delve"
	"nvim-go/ctx"

	"github.com/google/gops/agent"
	"github.com/neovim/go-client/nvim/plugin"
)

var (
	debug = os.Getenv("NVIM_GO_DEBUG")
	pprof = os.Getenv("NVIM_GO_PPROF")
)

func main() {
	register := func(p *plugin.Plugin) error {
		log.SetFlags(log.Lshortfile)

		ctxt := ctx.NewContext()
		c := command.Register(p, ctxt)
		delve.Register(p, ctxt)
		autocmd.Register(p, ctxt, c)

		if len(debug) >= 1 {
			// starts the gops agent
			if err := agent.Listen(&agent.Options{NoShutdownCleanup: true}); err != nil {
				return err
			}

			if len(pprof) >= 1 {
				addr := "localhost:14715" // (n: 14)vim-(g: 7)(o: 15)
				log.Printf("Start the pprof debugging, listen at %s\n", addr)

				// enable the report of goroutine blocking events
				runtime.SetBlockProfileRate(1)
				go func() {
					log.Println(http.ListenAndServe(addr, nil))
				}()
			}
		}
		return nil
	}

	plugin.Main(register)
}
