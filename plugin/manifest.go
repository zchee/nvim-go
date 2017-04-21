// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	old      = flag.Bool("old", false, "display old file data")
	new      = flag.Bool("new", false, "display new file data")
	manifest = flag.Bool("manifest", false, "display new plugin manifest")
	write    = flag.Bool("w", false, "write plugin manifest to file instead of stdout")
)

var manifestRe = regexp.MustCompile(`(?s)call remote#host#RegisterPlugin.+\\ ]\)`)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-old|-new|-manifest|-w] plugin_name\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	// Required first arg of plugin_name
	pluginName := flag.Arg(0)
	if pluginName == "" || flag.NFlag() > 1 {
		flag.Usage()
		os.Exit(2)
	}

	// Search gb binary path
	gbBin, err := exec.LookPath("gb")
	if err != nil {
		err = fmt.Errorf("does not exists gb binary: %v", err)
		log.Fatal(err)
	}

	// Get the gb project directory root
	gbCmd := exec.Command(gbBin)
	gbCmd.Args = append(gbCmd.Args, "env", "GB_PROJECT_DIR")
	gbResult, err := gbCmd.Output()
	if err != nil {
		err = fmt.Errorf("cannot get gb project directory: %v", err)
		log.Fatal(err)
	}
	prjDir := strings.TrimSpace(string(gbResult))

	// Get new plugin manifest
	manifestsCmd := exec.Command(filepath.Join(prjDir, "bin", pluginName), "-manifest", pluginName)
	newManifest, err := manifestsCmd.Output()
	if err != nil {
		panic(err)
	}
	newManifest = bytes.TrimSuffix(newManifest, []byte{'\n'})

	if strings.Contains(pluginName, "-race") {
		pluginName = strings.TrimSuffix(pluginName, "-race")
	}
	// Get vim file information from the "./plugin" directory
	pluginFile, err := os.OpenFile(filepath.Join(prjDir, "plugin", pluginName+".vim"), os.O_RDWR, os.FileMode(0))
	if err != nil {
		panic(err)
	}
	defer pluginFile.Close()

	// Read plugin vim file
	oldData, err := ioutil.ReadAll(pluginFile)
	if err != nil {
		panic(err)
	}

	// Replace the old specs to the latest specs
	newData := manifestRe.ReplaceAll(oldData, newManifest)

	// Output result
	switch {
	case *old:
		fmt.Printf("%v", string(oldData))
	case *new:
		fmt.Printf("%v", string(newData))
	case *manifest:
		// Trim last newline for output to stdout
		fmt.Printf("%v", string(newManifest))
	case *write:
		data := bytes.TrimSpace(newData)
		if bytes.Contains(data, []byte("-race")) {
			data = bytes.Replace(data, []byte("-race"), nil, -1)
		}
		data = append(data, '\n')
		if _, err := pluginFile.WriteAt(data, 0); err != nil {
			panic(err)
		}
	}
}
