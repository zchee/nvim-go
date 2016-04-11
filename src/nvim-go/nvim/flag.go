package nvim

import (
	"strconv"

	"github.com/garyburd/neovim-go/vim"
)

var (
	flags              []string
	buildAuto          = "go#build#autobuild"
	vBuildNomodifiable interface{}
	buildDebug         = "go#debug"
	vBuildDebug        interface{}
)

func ParseFlags(v *vim.Vim) []string {
	var (
		result  interface{}
		results []string
	)

	flags = append(flags, buildAuto, buildDebug)

	for _, flag := range flags {
		v.Var(flag, &result)

		switch value := result.(type) {
		case string:
			results = append(results, value)
		case int64:
			results = append(results, strconv.FormatInt(value, 10))
		case bool:
			results = append(results, strconv.FormatBool(value))
		default:
			Echoerr(v, "Not support interface type")
		}
	}

	return results
}

func isEmptyFlag(v *vim.Vim, flag interface{}) bool {
	switch v := flag.(type) {
	case string:
		if v != "" {
			return true
		} else {
			return false
		}
	case int64:
		if v > 0 {
			return true
		} else {
			return false
		}
	case bool:
		if !v {
			return true
		} else {
			return false
		}
	}
	return false
}
