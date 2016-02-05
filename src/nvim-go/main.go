// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command nvimgo is a Neovim remote plogin.
package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/garyburd/neovim-go/vim/plugin"

	_ "nvim-go/commands"
)

func init() {
	// logrus instead of stdlib log
	// neovim-go hijacked log format
	if lf := os.Getenv("NEOVIM_GO_LOG_FILE"); lf != "" {
		f, err := os.OpenFile(lf, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(f)
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	plugin.Main()
}
