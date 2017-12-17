// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(level string, debug bool, opts ...zap.Option) *zap.Logger {
	var cfg zap.Config
	if !debug {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		opts = append(opts, zap.AddCaller())
	}
	cfg.EncoderConfig.EncodeTime = nil

	if level != "" {
		var lv zapcore.Level
		if err := lv.UnmarshalText([]byte(level)); err != nil {
			panic(fmt.Sprintf("unknown zap log level: %v", level))
		}
		cfg.Level.SetLevel(lv)
	}

	zapLogger, err := cfg.Build(opts...)
	if err != nil {
		panic(err)
	}

	return zapLogger
}
