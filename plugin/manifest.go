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

	"github.com/pkg/errors"
)

var (
	old      = flag.Bool("old", false, "display old file data to stdout")
	new      = flag.Bool("new", false, "display new file data to stdout")
	manifest = flag.Bool("manifest", false, "display new plugin manifest to stdout")
	write    = flag.Bool("w", false, "write specs to file instead of stdout")
)

func main() {
	// Define usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-old|-new|-manifest|-w] plugin_name\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Required first arg of plugin_name
	pluginName := flag.Arg(0)
	if pluginName == "" || flag.NFlag() > 1 {
		flag.Usage()
		os.Exit(2)
	}

	// Search gb binary path
	gbBin, err := exec.LookPath("gb")
	if err != nil {
		err = errors.Wrap(err, "does not exists gb binary")
		log.Fatal(err)
	}

	// Get the gb project directory root
	gbCmd := exec.Command(gbBin)
	gbCmd.Args = append(gbCmd.Args, "env", "GB_PROJECT_DIR")
	gbResult, err := gbCmd.Output()
	if err != nil {
		err = errors.Wrap(err, "cannot get gb project directory")
		log.Fatal(err)
	}
	prjDir := strings.TrimSpace(string(gbResult))

	// Get new plugin manifest
	manifestsCmd := exec.Command(filepath.Join(prjDir, "bin", pluginName), "-manifest", pluginName)
	newManifest, err := manifestsCmd.Output()
	if err != nil {
		panic(err)
	}
	newManifest = append(newManifest, '\n')

	// Get vim file information from the "./plugin" directory
	plugFile, err := os.OpenFile(filepath.Join(prjDir, "plugin", pluginName+".vim"), os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer plugFile.Close()

	// Read plugin vim file
	oldData, err := ioutil.ReadAll(plugFile)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`(?s)call remote#host#RegisterPlugin.+`)
	// Replace the old specs to the latest specs
	newData := re.ReplaceAll(oldData, newManifest)

	// Output result
	switch {
	case *old:
		fmt.Printf("%v", string(oldData))
		return
	case *new:
		fmt.Printf("%v", string(newData))
		return
	case *manifest:
		// Trim last newline for output to stdout
		fmt.Printf("%v", string(newManifest[:len(newManifest)-1]))
		return
	case *write:
		data := bytes.TrimSpace(newData)
		data = append(data, byte('\n'))
		if _, err := plugFile.WriteAt(data, 0); err != nil {
			panic(err)
		}
		return
	}
}
