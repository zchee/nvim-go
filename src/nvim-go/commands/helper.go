package commands

import "strings"

// ToPascalCase convert s to PascalCase.
// This function assumes that the character of the beginning is a-z.
func ToPascalCase(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

// ToMixedCase convert s to mixedCase.
// This function assumes that the character of the beginning is A-Z.
func ToMixedCase(s string) string { return strings.ToLower(s[:1]) + s[1:] }
