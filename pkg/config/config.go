// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	xdgbasedir "github.com/zchee/go-xdgbasedir"
	yaml "gopkg.in/yaml.v2"
)

var (
	mkdirOnce  sync.Once
	ConfigHome = filepath.Join(xdgbasedir.ConfigHome(), "nvim-go")
	ConfigFile = filepath.Join(ConfigHome, "config.yml")
)

func CreateConfigHome() error {
	var err error
	mkdirOnce.Do(func() {
		if _, e := os.Stat(ConfigHome); e != nil && os.IsNotExist(e) {
			if e := os.MkdirAll(ConfigHome, 0700); e != nil {
				err = e
				return
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
		if err := CreateConfigHome(); err != nil {
			return nil, err
		}
		f, err = os.Create(ConfigFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create %s", ConfigFile)
		}
	}

	return f, nil
}

func Read() (*Config, error) {
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

func Merge(cfg, cfg2 *Config) *Config {
	// early return
	if cfg2 == nil {
		return cfg
	}

	if cfg2.Global != nil {
		if cfg.Global.ErrorListType != cfg2.Global.ErrorListType {
			cfg.Global.ErrorListType = cfg2.Global.ErrorListType
		}
	}

	if cfg2.Build != nil {
		if itob(cfg.Build.Appengine) != itob(cfg2.Build.Appengine) {
			cfg.Build.Appengine = cfg2.Build.Appengine
		}
		if itob(cfg.Build.Autosave) != itob(cfg2.Build.Autosave) {
			cfg.Build.Autosave = cfg2.Build.Autosave
		}
		if itob(cfg.Build.Autosave) != itob(cfg2.Build.Autosave) {
			cfg.Build.Autosave = cfg2.Build.Autosave
		}
		if strings.EqualFold(strings.Join(cfg.Build.Flags, ""), strings.Join(cfg2.Build.Flags, "")) {
			cfg.Build.Flags = cfg2.Build.Flags
		}
		if itob(cfg.Build.Force) != itob(cfg2.Build.Force) {
			cfg.Build.Force = cfg2.Build.Force
		}
	}

	if cfg2.Cover != nil {
		if strings.EqualFold(strings.Join(cfg.Cover.Flags, ""), strings.Join(cfg2.Cover.Flags, "")) {
			cfg.Cover.Flags = cfg2.Cover.Flags
		}
		if cfg.Cover.Mode != cfg2.Cover.Mode {
			cfg.Cover.Mode = cfg2.Cover.Mode
		}
	}

	if cfg2.Fmt != nil {
		if itob(cfg.Fmt.Autosave) != itob(cfg2.Fmt.Autosave) {
			cfg2.Fmt.Autosave = cfg2.Fmt.Autosave
		}
		if cfg.Fmt.Mode != cfg2.Fmt.Mode {
			cfg.Fmt.Mode = cfg2.Fmt.Mode
		}
	}

	if cfg2.Generate != nil {
		if itob(cfg.Generate.TestAllFuncs) != itob(cfg2.Generate.TestAllFuncs) {
			cfg.Generate.TestAllFuncs = cfg2.Generate.TestAllFuncs
		}
		if cfg.Generate.TestExclFuncs != cfg2.Generate.TestExclFuncs {
			cfg.Generate.TestExclFuncs = cfg2.Generate.TestExclFuncs
		}
		if itob(cfg.Generate.TestExportedFuncs) != itob(cfg2.Generate.TestExportedFuncs) {
			cfg.Generate.TestExportedFuncs = cfg2.Generate.TestExportedFuncs
		}
		if itob(cfg.Generate.TestSubTest) != itob(cfg2.Generate.TestSubTest) {
			cfg.Generate.TestSubTest = cfg2.Generate.TestSubTest
		}
	}

	if cfg2.Guru != nil {
		if itob(cfg.Guru.JumpFirst) != itob(cfg2.Guru.JumpFirst) {
			cfg.Guru.JumpFirst = cfg2.Guru.JumpFirst
		}
		if cfg2.Guru.KeepCursor != nil {
			cfg.Guru.KeepCursor = cfg2.Guru.KeepCursor
		}
		if itob(cfg.Guru.Reflection) != itob(cfg2.Guru.Reflection) {
			cfg.Guru.Reflection = cfg2.Guru.Reflection
		}
	}

	if cfg2.Iferr != nil {
		if itob(cfg.Iferr.Autosave) != itob(cfg2.Iferr.Autosave) {
			cfg.Iferr.Autosave = cfg.Iferr.Autosave
		}
	}

	if cfg2.Lint != nil {
		if itob(cfg.Lint.GoVetAutosave) != itob(cfg2.Lint.GoVetAutosave) {
			cfg.Lint.GoVetAutosave = cfg2.Lint.GoVetAutosave
		}
		if strings.EqualFold(strings.Join(cfg.Lint.GoVetFlags, ""), strings.Join(cfg2.Lint.GoVetFlags, "")) {
			cfg.Lint.GoVetFlags = cfg2.Lint.GoVetFlags
		}
		if strings.EqualFold(strings.Join(cfg.Lint.GoVetIgnore, ""), strings.Join(cfg2.Lint.GoVetIgnore, "")) {
			cfg.Lint.GoVetIgnore = cfg2.Lint.GoVetIgnore
		}
		if itob(cfg.Lint.GolintAutosave) != itob(cfg2.Lint.GolintAutosave) {
			cfg.Lint.GolintAutosave = cfg2.Lint.GolintAutosave
		}
		if strings.EqualFold(strings.Join(cfg.Lint.GolintIgnore, ""), strings.Join(cfg2.Lint.GolintIgnore, "")) {
			cfg.Lint.GolintIgnore = cfg2.Lint.GolintIgnore
		}
		if cfg.Lint.GolintMinConfidence != cfg2.Lint.GolintMinConfidence {
			cfg.Lint.GolintMinConfidence = cfg2.Lint.GolintMinConfidence
		}
		if cfg.Lint.GolintMode != cfg2.Lint.GolintMode {
			cfg.Lint.GolintMode = cfg2.Lint.GolintMode
		}
		if itob(cfg.Lint.MetalinterAutosave) != itob(cfg2.Lint.MetalinterAutosave) {
			cfg.Lint.MetalinterAutosave = cfg2.Lint.MetalinterAutosave
		}
		if strings.EqualFold(strings.Join(cfg.Lint.MetalinterAutosaveTools, ""), strings.Join(cfg2.Lint.MetalinterAutosaveTools, "")) {
			cfg.Lint.MetalinterAutosaveTools = cfg2.Lint.MetalinterAutosaveTools
		}
		if cfg.Lint.MetalinterDeadline != cfg2.Lint.MetalinterDeadline {
			cfg.Lint.MetalinterDeadline = cfg2.Lint.MetalinterDeadline
		}
		if strings.EqualFold(strings.Join(cfg.Lint.MetalinterSkipDir, ""), strings.Join(cfg2.Lint.MetalinterSkipDir, "")) {
			cfg.Lint.MetalinterSkipDir = cfg2.Lint.MetalinterSkipDir
		}
		if strings.EqualFold(strings.Join(cfg.Lint.MetalinterTools, ""), strings.Join(cfg2.Lint.MetalinterTools, "")) {
			cfg.Lint.MetalinterTools = cfg2.Lint.MetalinterTools
		}
	}

	if cfg2.Rename != nil {
		if itob(cfg.Rename.Prefill) != itob(cfg2.Rename.Prefill) {
			cfg.Rename.Prefill = cfg2.Rename.Prefill
		}
	}

	if cfg2.Terminal != nil {
		if cfg.Terminal.Height != cfg2.Terminal.Height {
			cfg.Terminal.Height = cfg2.Terminal.Height
		}
		if cfg.Terminal.Mode != cfg2.Terminal.Mode {
			cfg.Terminal.Mode = cfg2.Terminal.Mode
		}
		if cfg.Terminal.Position != cfg2.Terminal.Position {
			cfg.Terminal.Position = cfg2.Terminal.Position
		}
		if itob(cfg.Terminal.StopInsert) != itob(cfg.Terminal.StopInsert) {
			cfg.Terminal.StopInsert = cfg2.Terminal.StopInsert
		}
		if cfg.Terminal.Width != cfg2.Terminal.Width {
			cfg.Terminal.Width = cfg2.Terminal.Width
		}
	}

	if cfg2.Test != nil {
		if itob(cfg.Test.AllPackage) != itob(cfg2.Test.AllPackage) {
			cfg.Test.AllPackage = cfg2.Test.AllPackage
		}
		if itob(cfg.Test.Autosave) != itob(cfg2.Test.Autosave) {
			cfg.Test.Autosave = cfg2.Test.Autosave
		}
		if strings.EqualFold(strings.Join(cfg.Test.Flags, ""), strings.Join(cfg2.Test.Flags, "")) {
			cfg.Test.Flags = cfg2.Test.Flags
		}
	}

	if cfg2.Debug != nil {
		if itob(cfg.Debug.Enable) != itob(cfg2.Debug.Enable) {
			cfg.Debug.Enable = cfg2.Debug.Enable
		}
		if itob(cfg.Debug.Pprof) != itob(cfg2.Debug.Pprof) {
			cfg.Debug.Pprof = cfg2.Debug.Pprof
		}
	}

	return cfg
}
