// Copyright 2015 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program prints Neovim's API info as JSON.
//
// A snapshot of the output from this program is checked into api.json.
// Compare api.json to the output from this program to discover changes to
// Neovim's API.

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"

	"github.com/neovim-go/msgpack"
)

func main() {
	log.SetFlags(0)

	output, err := exec.Command("nvim", "--api-info").Output()
	if err != nil {
		log.Fatalf("error getting API info: %v", err)
	}

	var v interface{}
	if err := msgpack.NewDecoder(bytes.NewReader(output)).Decode(&v); err != nil {
		log.Fatalf("error parsing msppack: %v", err)
	}

	p, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(append(p, '\n'))
}
