package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	old   = flag.Bool("old", false, "display old file data to stdout")
	new   = flag.Bool("new", false, "display new file data to stdout")
	spec  = flag.Bool("specs", false, "display latest specs to stdout")
	write = flag.Bool("w", false, "write specs to file instead of stdout")
)

func main() {
	// Define usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-old|-new|-specs|-w] plugin_name\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Required first arg of plugin_name
	pluginName := flag.Arg(0)
	if pluginName == "" || flag.NFlag() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Get gb project directory
	gbCmd := exec.Command("gb", "env", "GB_PROJECT_DIR")
	gbResult, err := gbCmd.Output()
	if err != nil {
		panic(err)
	}
	gbCmd.Run()
	projectDir := strings.TrimSpace(string(gbResult))

	// Get latest plugin specs
	specsCmd := exec.Command(projectDir+"/bin/"+pluginName, "-specs")
	newSpecs, err := specsCmd.Output()
	if err != nil {
		panic(err)
	}
	specsCmd.Run()
	newSpecs = append(newSpecs, byte('\n'))

	// Get vim file information from the `plugin` directory
	plugFile, err := os.OpenFile(projectDir+"/plugin/"+pluginName+".vim", os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	defer plugFile.Close()

	// Read plugin vim file
	oldData, err := ioutil.ReadAll(plugFile)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`(?s)let s:specs =.+]\n+`)
	// Replace the old specs to the latest specs
	newData := re.ReplaceAll(oldData, newSpecs)

	// Output result
	switch {
	case *old:
		fmt.Printf("%v", string(oldData))
		return
	case *new:
		fmt.Printf("%v", string(newData))
		return
	case *spec:
		// Trim last newline for output to stdout
		fmt.Printf("%v", string(newSpecs[:len(newSpecs)-1]))
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
