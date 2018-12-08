// Copyright 2018 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strutil

import (
	"strings"
)

// convert converts a s to (lower)CamelCase.
func convert(s string, initUpper bool) (n string) {
	s = strings.Trim(s, " ")
	capNext := initUpper

	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			if capNext {
				n += string(r ^ 32)
			} else {
				n += string(r)
			}
		case r >= 'A' && r <= 'Z':
			n += string(r)
		case r >= '0' && r <= '9':
			n += string(r)
		}

		if r == '_' || r == ' ' || r == '-' || r >= '0' && r <= '9' {
			capNext = true
			continue
		}
		capNext = false
	}

	return n
}

// ToCamelCase converts a s to CamelCase.
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}

	return convert(s, true)
}

// ToLowerCamelCase converts a s to lower CamelCase.
func ToLowerCamelCase(s string) string {
	if s == "" {
		return s
	}

	if r := rune(s[0]); r >= 'A' && r <= 'Z' {
		s = string(r^32) + s[1:]
	}

	return convert(s, false)
}

// ToSnakeCase converts a s to snake_case.
func ToSnakeCase(s string) string {
	return ToDelimited(s, '_')
}

// ToScreamingSnakeCase converts a s to SCREAMING_SNAKE_CASE.
func ToScreamingSnakeCase(s string) string {
	return ToScreamingDelimited(s, '_', true)
}

// ToKebab converts a s to kebab-case.
func ToKebab(s string) string {
	return ToDelimited(s, '-')
}

// ToScreamingKebab converts a s to SCREAMING-KEBAB-CASE.
func ToScreamingKebab(s string) string {
	return ToScreamingDelimited(s, '-', true)
}

// ToDelimited converts a s to delimited.snake.case (in this case `del = '.'`).
func ToDelimited(s string, del uint8) string {
	return ToScreamingDelimited(s, del, false)
}

// ToScreamingDelimited converts a s to SCREAMING.DELIMITED.SNAKE.CASE
// in this case `del = '.'; screaming = true`) or delimited.snake.case (in this case `del = '.'; screaming = false`.
func ToScreamingDelimited(s string, del uint8, screaming bool) (n string) {
	s = strings.Trim(s, " ")

	for i, r := range s {
		nextCaseIsChanged := false
		if i+1 < len(s) {
			next := s[i+1]
			if (r >= 'A' && r <= 'Z' && next >= 'a' && next <= 'z') || (r >= 'a' && r <= 'z' && next >= 'A' && next <= 'Z') {
				nextCaseIsChanged = true
			}
		}

		switch {
		case i > 0 && n[len(n)-1] != del && nextCaseIsChanged:
			switch {
			case r >= 'a' && r <= 'z':
				n += string(r) + string(del)
			case r >= 'A' && r <= 'Z':
				n += string(del) + string(r)
			}
		case r == ' ' || r == '_' || r == '-':
			n += string(del)
		default:
			n = n + string(r)
		}
	}

	if screaming {
		var s string
		for _, r := range n {
			if r >= 'a' && r <= 'z' {
				s += string(r ^ 32)
				continue
			}
			s += string(r)
		}
		n = s
	} else {
		var s string
		for _, r := range n {
			if r >= 'A' && r <= 'Z' {
				s += string(r ^ 32)
				continue
			}
			s += string(r)
		}
		n = s
	}

	return n
}
