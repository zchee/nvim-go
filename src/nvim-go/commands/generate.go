package commands

import (
	"bufio"
	"nvim-go/nvim"
	"os"

	"github.com/cweill/gotests/gotests/process"
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	plugin.HandleCommand("GoGenerateTest", &plugin.CommandOptions{NArgs: "*", Complete: "file"}, cmdGenerateTest)
}

func cmdGenerateTest(v *vim.Vim, files []string) {
	go GenerateTest(v, files)
}

func GenerateTest(v *vim.Vim, files []string) error {
	b, err := v.CurrentBuffer()
	if err != nil {
		return nvim.Echoerr(v, "GoGenerateTest: %v", err)
	}

	if len(files) == 0 {
		f, err := v.BufferName(b)
		if err != nil {
			return nvim.Echoerr(v, "GoGenerateTest: %v", err)
		}
		files = append(files, f)
	}

	var opt = process.Options{
		AllFuncs:    true,
		WriteOutput: true,
		PrintInputs: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	process.Run(w, files, &opt)

	w.Close()
	os.Stdout = oldStdout

	var out string
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		out += scan.Text() + "\n"
	}

	return nvim.EchoRaw(v, out)
}
