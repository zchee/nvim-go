// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
)

// env represents a environment variabels for nvim-go.
type env struct {
	GCPProjectID string `envconfig:"GCP_PROJECT_ID"`
	Debug        bool   `envconfig:"DEBUG"`
	LogLevel     string `envconfig:"LOG_LEVEL"`
}

func (e env) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("gcp_project_id", e.GCPProjectID)
	enc.AddBool("debug", e.Debug)
	enc.AddString("log_level", e.LogLevel)

	return nil
}

// e is the global env variable
var e env

// envOnce for run Process once.
var envOnce sync.Once

// Process populates the specified struct based on environment variables.
func Process() env {
	envOnce.Do(func() {
		envconfig.MustProcess("NVIM_GO", &e)
	})

	return e
}

func GCPProjectID() string {
	_ = Process()
	return e.GCPProjectID
}

func HasGCPProjectID() bool {
	_ = Process()
	return e.GCPProjectID != ""
}

func IsDebug() bool {
	_ = Process()
	return e.Debug
}
