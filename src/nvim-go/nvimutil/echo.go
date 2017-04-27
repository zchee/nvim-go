// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nvimutil

import (
	"fmt"
	"log"

	"nvim-go/config"

	"github.com/neovim/go-client/nvim"
	"github.com/pkg/errors"
)

var (
	// ErrorColor highlight error message use Identifier syntax color.
	ErrorColor = "Identifier"
	// ProgressColor highlight progress message use Identifier syntax color.
	ProgressColor = "Identifier"
	// SuccessColor highlight success message use Identifier syntax color.
	SuccessColor = "Function"
)

// Echo provide the vim 'echo' command.
func Echo(v *nvim.Nvim, format string, a ...interface{}) error {
	v.Command("redraw")
	return v.Command("echo '" + fmt.Sprintf(format, a...) + "'")
}

// EchoRaw provide the raw output vim 'echo' command.
func EchoRaw(v *nvim.Nvim, a string) error {
	v.Command("redraw")
	return v.Command("echo \"" + a + "\"")
}

// Echomsg provide the vim 'echomsg' command.
func Echomsg(v *nvim.Nvim, a ...interface{}) error {
	return v.WriteOut(fmt.Sprintln(a...))
}

// Echoerr provide the vim 'echoerr' command.
func Echoerr(v *nvim.Nvim, format string, a ...interface{}) error {
	return v.WritelnErr(fmt.Sprintf(format, a...))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ErrorWrap splits the errors.Wrap's cause and error messages,
// and provide the vim 'echo' message with 'echohl' highlighting to cause text.
func ErrorWrap(v *nvim.Nvim, err error) error {
	if err == nil {
		return nil
	}

	var funcName string
	if err, ok := err.(stackTracer); ok {
		st := err.StackTrace()
		// "%n" verb is function name
		funcName = fmt.Sprintf("%n", st[0])
		if config.DebugEnable {
			log.Printf("Error stack%+v", st[:])
		}
	}
	// fallback use plugin name
	if funcName == "" {
		funcName = "nvim-go"
	}

	return v.WritelnErr(fmt.Sprintf("%s: %s", funcName, err))
}

// EchohlBefore provide the vim 'echo' command with the 'echohl' highlighting prefix text.
func EchohlBefore(v *nvim.Nvim, prefix string, highlight string, format string, a ...interface{}) error {
	v.Command("redraw")
	suffix := "\" | echohl None | echon \""
	if prefix != "" {
		suffix += ": "
	}
	return v.Command("echohl " + highlight + " | echo \"" + prefix + suffix + fmt.Sprintf(format, a...) + "\" | echohl None")
}

// EchohlAfter provide the vim 'echo' command with the 'echohl' highlighting message text.
func EchohlAfter(v *nvim.Nvim, prefix string, highlight string, format string, a ...interface{}) error {
	v.Command("redraw")
	if prefix != "" {
		prefix += ": "
	}
	return v.Command("echo \"" + prefix + "\" | echohl " + highlight + " | echon \"" + fmt.Sprintf(format, a...) + "\" | echohl None")
}

// EchoProgress displays a command progress message to echo area.
func EchoProgress(v *nvim.Nvim, prefix, format string, a ...interface{}) error {
	v.Command("redraw")
	msg := fmt.Sprintf(format, a...)
	return v.Command(fmt.Sprintf("echo \"%s: \" | echohl %s | echon \"%s ...\" | echohl None", prefix, ProgressColor, msg))
}

// EchoSuccess displays the success of the command to echo area.
func EchoSuccess(v *nvim.Nvim, prefix string, msg string) error {
	v.Command("redraw")
	if msg != "" {
		msg = " | " + msg
	}
	return v.Command(fmt.Sprintf("echo \"%s: \" | echohl %s | echon 'SUCCESS' | echohl None | echon '%s' | echohl None", prefix, SuccessColor, msg))
}

// ReportError output of the accumulated errors report.
// TODO(zchee): research vim.ReportError behavior
// Why it does not immediately display error?
// func ReportError(v *vim.Nvim, format string, a ...interface{}) error {
// 	return v.ReportError(fmt.Sprintf(format, a...))
// }

// ClearMsg cleanups the echo area.
func ClearMsg(v *nvim.Nvim) error {
	return v.Command("echon")
}
