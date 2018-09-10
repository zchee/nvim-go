package structtag

import (
	"errors"
	"reflect"
	"strconv"
)

var notFound = errors.New(`specified tag not found`)

// BoolValue looks up tag `n`, and returns its value after
// parsing it as boolean using strconv.ParseBool.
func BoolValue(t reflect.StructTag, n string) (bool, error) {
	s, ok := Lookup(t, n)
	if !ok {
		return false, notFound
	}

	return strconv.ParseBool(s)
}
