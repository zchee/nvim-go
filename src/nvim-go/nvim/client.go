package nvim

import (
	"log"
	"net"
	"nvim-go/config"
	"os"
	"regexp"

	"github.com/neovim-go/vim"
)

// NewSocketClient creates the Neovim client over the socket session.
func NewSocketClient() *vim.Vim {
	var (
		v   *vim.Vim
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
func NewEmbedClient(args []string, dir string, env []string) *vim.Vim {
	options := &vim.EmbedOptions{
		Args: args,
		Dir:  dir,
		Env:  env,
	}

	v, err := vim.NewEmbedded(options)
	if err != nil {
		log.Println(err)
		return nil
	}

	return v
}

// NewStdioClient creates the Neovim client over the stdio.
func NewStdioClient() *vim.Vim {
	v, err := vim.New(os.Stdin, os.Stdout, os.Stdout, log.Printf)
	if err != nil {
		log.Println(err)
		return nil
	}

	go v.Serve()

	return v
}

var tcpAddrRe = regexp.MustCompile(`:\d+$`)

func dialNvim(addr string) (*vim.Vim, error) {
	network := "unix"
	if tcpAddrRe.MatchString(addr) {
		network = "tcp"
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	v, err := vim.New(conn, conn, conn, log.Printf)
	if err != nil {
		return nil, err
	}
	go v.Serve()

	return v, nil
}
