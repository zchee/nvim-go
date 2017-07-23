// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
)

var (
	mkdirOnce  sync.Once
	ConfigHome = filepath.Join(xdgbasedir.ConfigHome(), "nvim-go")
	ConfigFile = filepath.Join(ConfigHome, "config.yml")
)

func CreateConfigDir() error {
	var err error
	mkdirOnce.Do(func() {
		if _, e := os.Stat(ConfigHome); e != nil && os.IsNotExist(e) {
			if e := os.MkdirAll(ConfigHome, 0700); e != nil {
				err = e
			}
			err = e
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func open() (*os.File, error) {
	f, err := os.Open(ConfigFile)
	if err != nil && os.IsNotExist(err) {
		f, err = os.Create(ConfigFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create %s", ConfigFile)
		}
	}

	return f, nil
}

func ReadConfig() (*Config, error) {
	f, err := open()
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read %s", f.Name())
	}

	cfg := new(Config)
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal %s", f.Name())
	}

	return cfg, nil
}
