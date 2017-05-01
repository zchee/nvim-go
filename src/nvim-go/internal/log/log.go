// Copyright 2017 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"log"
	"os"
)

var (
	logger      = log.New(os.Stderr, "", log.Lshortfile)
	debug       = os.Getenv("NVIM_GO_DEBUG")
	debugLogger = log.New(os.Stderr, "DEBUG: ", log.Lshortfile)
)

func Print(v ...interface{}) {
	logger.Output(2, fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	logger.Output(2, fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	logger.Output(2, fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	logger.Output(2, "FATAL: "+fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	logger.Output(2, "FATAL: "+fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	logger.Output(2, "FATAL: "+fmt.Sprintln(v...))
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.Output(2, "PANIC: "+s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.Output(2, "PANIC: "+s)
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	logger.Output(2, "PANIC: "+s)
	panic(s)
}

func Debug(v ...interface{}) {
	if len(debug) == 0 {
		return
	}
	debugLogger.Output(2, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	if len(debug) == 0 {
		return
	}
	debugLogger.Output(2, fmt.Sprintf(format, v...))
}

func Debugln(v ...interface{}) {
	if len(debug) == 0 {
		return
	}
	debugLogger.Output(2, fmt.Sprintln(v...))
}
