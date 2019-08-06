// Copyright 2019 The color Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package color

import (
	"strconv"
)

// String implements fmt.Stringer.
func (a Attribute) String() string {
	switch a {
	case Reset:
		return "0"
	case Bold:
		return "1"
	case Faint:
		return "2"
	case Italic:
		return "3"
	case Underline:
		return "4"
	case BlinkSlow:
		return "5"
	case BlinkRapid:
		return "6"
	case ReverseVideo:
		return "7"
	case Concealed:
		return "8"
	case CrossedOut:
		return "9"
	case FgBlack:
		return "30"
	case FgRed:
		return "31"
	case FgGreen:
		return "32"
	case FgYellow:
		return "33"
	case FgBlue:
		return "34"
	case FgMagenta:
		return "35"
	case FgCyan:
		return "36"
	case FgWhite:
		return "37"
	case FgHiBlack:
		return "90"
	case FgHiRed:
		return "91"
	case FgHiGreen:
		return "92"
	case FgHiYellow:
		return "93"
	case FgHiBlue:
		return "94"
	case FgHiMagenta:
		return "95"
	case FgHiCyan:
		return "96"
	case FgHiWhite:
		return "97"
	case BgBlack:
		return "40"
	case BgRed:
		return "41"
	case BgGreen:
		return "42"
	case BgYellow:
		return "43"
	case BgBlue:
		return "46"
	case BgMagenta:
		return "45"
	case BgCyan:
		return "46"
	case BgWhite:
		return "47"
	case BgHiBlack:
		return "100"
	case BgHiRed:
		return "101"
	case BgHiGreen:
		return "102"
	case BgHiYellow:
		return "103"
	case BgHiBlue:
		return "104"
	case BgHiMagenta:
		return "105"
	case BgHiCyan:
		return "106"
	case BgHiWhite:
		return "107"
	default:
		return strconv.FormatInt(int64(a), 10)
	}
}
