// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"sync"
	"sync/atomic"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
)

// Env represents a environment variabels for nvim-go.
type Env struct {
	EnabledStackdriverProfiler   bool   `envconfig:"ENABLED_STACKDRIVER_PROFILER" default:"false"`
	EnabledStackdriverExporter   bool   `envconfig:"ENABLED_STACKDRIVER_EXPORTER" default:"false"`
	GCPProjectID                 string `envconfig:"GOOGLE_CLOUD_PROJECT"`
	GoogleApplicationCredentials string `envconfig:"GOOGLE_APPLICATION_CREDENTIALS"`
	EnabledDatadogExporter       bool   `envconfig:"ENABLED_DATADOG_EXPORTER" default:"false"`
	DatadogAgentHostname         string `envconfig:"DATADOG_AGENT_HOSTNAME" default:"127.0.0.1"`
	Debug                        bool   `envconfig:"DEBUG" default:"false"`
	LogLevel                     string `envconfig:"LOG_LEVEL" default:"info"`
}

func (e Env) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("gcp_project_id", e.GCPProjectID)
	enc.AddString("google_application_credentials", e.GoogleApplicationCredentials)
	enc.AddString("datadog_agent_hostname", e.DatadogAgentHostname)
	enc.AddBool("debug", e.Debug)
	enc.AddString("log_level", e.LogLevel)

	return nil
}

var (
	// e is the global env variable
	e Env

	// envOnce for run Process once.
	envOnce sync.Once

	done uint32
)

// Process populates the specified struct based on environment variables.
func Process() Env {
	envOnce.Do(func() {
		envconfig.MustProcess("NVIM_GO", &e)
		atomic.StoreUint32(&done, 1)
	})

	return e
}

// GCPProjectID return the GCP Project ID from the $GOOGLE_CLOUD_PROJECT environment variable.
func GCPProjectID() string {
	if atomic.LoadUint32(&done) == 0 {
		_ = Process()
	}

	return e.GCPProjectID
}

// HasGCPProjectID reports whether the has $GOOGLE_CLOUD_PROJECT environment variable and return the GCP Project ID.
func HasGCPProjectID() (string, bool) {
	if atomic.LoadUint32(&done) == 0 {
		_ = Process()
	}

	return e.GCPProjectID, e.GCPProjectID != ""
}

// IsDebug reports whether the debug environment.
func IsDebug() bool {
	if atomic.LoadUint32(&done) == 0 {
		_ = Process()
	}

	return e.Debug
}
