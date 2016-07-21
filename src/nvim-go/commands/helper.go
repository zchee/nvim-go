package commands

import (
	"reflect"
	"strings"
	"unsafe"
)

// ToPascalCase convert s to PascalCase.
// This function assumes that the character of the beginning is a-z.
func ToPascalCase(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

// ToMixedCase convert s to mixedCase.
// This function assumes that the character of the beginning is A-Z.
func ToMixedCase(s string) string { return strings.ToLower(s[:1]) + s[1:] }

// StringToByteslice convert string to byte slice use unsafe.
// https://gist.github.com/dgryski/65d632958e4d88c7f79aaa7e1d2b10c0
func StringToByteslice(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	bp := (*[]byte)(unsafe.Pointer(bh))
	return *bp
}
