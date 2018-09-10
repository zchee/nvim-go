// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"sync"

	envpkg "github.com/lestrrat-go/config/env"
	"go.uber.org/zap/zapcore"
)

// e is the global env variable
var e env

// env represents a environment variabels with lestrrat-go/config/env struct tag.
type env struct {
	GCPProjectID string `envconfig:"GCP_PROJECT_ID"`
	Debug        bool   `envconfig:"DEBUG"`
}

func (e env) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("gcp_project_id", e.GCPProjectID)
	enc.AddBool("debug", e.Debug)

	return nil
}

// envOnce for run Process once.
var envOnce sync.Once

// Process populates the specified struct based on environment variables.
func Process() (e env, err error) {
	envOnce.Do(func() {
		err = envpkg.NewDecoder(envpkg.System).Prefix("NVIM_GO").Decode(&e)
	})
	return e, err
}

func GCPProjectID() string {
	_, _ = Process()
	return e.GCPProjectID
}

func IsDebug() bool {
	_, _ = Process()
	return e.Debug
}
