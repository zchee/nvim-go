// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	xdgbasedir "github.com/zchee/go-xdgbasedir"

	"github.com/zchee/nvim-go/pkg/fs"
	"github.com/zchee/nvim-go/pkg/nvimutil"
)

// CmdConfigEval represents for the eval of GoConfig command.
type CmdConfigEval struct {
	Cwd  string `msgpack:",array"`
	File string
}

var (
	configDir  = filepath.Join(xdgbasedir.ConfigHome(), "nvim-go")
	configFile = filepath.Join(configDir, "config.yml")
)

func (c *Command) cmdConfig(ctx context.Context, args []string, eval *CmdConfigEval) {
	if err := fs.Mkdir(configDir, 0700); err != nil {
		nvimutil.Echoerr(c.Nvim, "%v", err)
		return
	}
	if err := fs.Create(configFile); err != nil {
		nvimutil.Echoerr(c.Nvim, "%v", err)
		return
	}

	go func() {
		c.errs.Delete("config")

		err := c.Config(ctx, args, eval)
		switch e := err.(type) {
		case error:
			nvimutil.ErrorWrap(c.Nvim, e)
		case []*nvim.QuickfixError:
			c.errs.Store("Build", e)
			errlist := make(map[string][]*nvim.QuickfixError)
			c.errs.Range(func(ki, vi interface{}) bool {
				k, v := ki.(string), vi.([]*nvim.QuickfixError)
				errlist[k] = append(errlist[k], v...)
				return true
			})
			nvimutil.ErrorList(c.Nvim, errlist, true)
		}
	}()
}

// Config configs nvim-go plugin specific config.
// Such as enable appengine mode, set GOPATH for project or etc.
func (c *Command) Config(ctx context.Context, args []string, eval *CmdConfigEval) interface{} {
	defer nvimutil.Profile(ctx, time.Now(), "GoConfig")

	if err := c.parseConfigCmd(args[0], args[1:], eval.Cwd); err != nil {
		return err
	}

	return nvimutil.EchoSuccess(c.Nvim, "GoConfig", fmt.Sprintf("Set %s to %s", args[0], args[1:]))
}

type Config struct {
	Project map[string]ProjectConfig `yaml:"project"`
}

type ProjectConfig struct {
	dir         string `yaml:"gopath"`
	GOPATH      string `yaml:"gopath"`
	IsAppengine bool   `yaml:"is_appengine"`
	IsGb        bool   `yaml:"is_not_gb"`
}

func (c *Config) Marshal() ([]byte, error) {
	fi, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	buf, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	out, err := yaml.Marshal(&buf)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Config) Unmarshal() (*Config, error) {
	fi, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	buf, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(buf, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Command) parseConfigCmd(cmd string, args []string, dir string) error {
	switch cmd {
	case "gopath":
		switch len(args) {
		case 0:
			return errors.New("gopath subcommand need GOPATH directory path")
		case 1:
			// nothing to do
		default:
			return errors.New("invalid arguments of gopath subcommand")
		}
		gopath := args[0]

		pjcfg := ProjectConfig{
			dir:    dir,
			GOPATH: gopath,
		}
		if err := writeConfig(pjcfg); err != nil {
			return err
		}

	}
	return nil
}

func writeConfig(pjcfg ProjectConfig) error {
	fi, err := os.OpenFile(configFile, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer fi.Close()

	cfg := make(map[string]ProjectConfig)
	cfg[fs.FindVCSRoot(pjcfg.dir)] = pjcfg
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "could not marshal to yaml")
	}
	if _, err := fi.Write(buf); err != nil {
		return errors.Wrapf(err, "could not write yaml data to %s", configFile)
	}

	return nil
}

func readConfig() (*Config, error) {
	fi, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	buf, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
