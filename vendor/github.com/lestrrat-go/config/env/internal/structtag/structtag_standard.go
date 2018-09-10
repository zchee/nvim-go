// +build !appengine

package structtag

import "reflect"

func Lookup(t reflect.StructTag, s string) (string, bool) {
	return t.Lookup(s)
}
