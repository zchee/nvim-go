// Copyright 2019 The color Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package color

func init() {
	m := make(map[AttributeHash]*Color, 32) // Total for loop is 8(Black~White) * 4({F,B}g{,Hi}) = 32

	for _, attrs := range [4][2]Attribute{
		{
			FgBlack,
			FgWhite,
		},
		{
			FgHiBlack,
			FgHiWhite,
		},
		{
			BgBlack,
			BgWhite,
		},
		{
			BgHiBlack,
			BgHiWhite,
		},
	} {
		start := attrs[0]
		end := attrs[1]
		for attr := start; attr < end; attr++ {
			m[hashAttributes(attr)] = &Color{params: []Attribute{attr}}
		}
	}

	colorCache.Put(m)
}
