// +build appengine

package structtag

import "reflect"

func Lookup(t reflect.StructTag, s string) (string, bool) {
	// for appengine, we consider an empty string as non-existant
	v := t.Get(s)
	return v, v != ""
}
