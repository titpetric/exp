package model

import (
	"strings"
)

// To type returns a sanitized type name.
func ToType(s string) (string, bool) {
	// variadic ...
	s = strings.TrimLeft(s, ".")
	// functions
	if strings.HasPrefix(s, "func") {
		return "<func>", false
	}
	// channels
	// implementation gap: anonymous structs
	if strings.Contains(s, " ") {
		ss := strings.SplitN(s, " ", 2)
		if strings.Contains(ss[0], "chan") {
			s = ss[1]
		}
	}
	// maps, slices
	for strings.Contains(s, "]") {
		// Handle:
		// - `[]` V
		// - `map[...]` V
		// - `.*]` V
		ss := strings.SplitN(s, "]", 2)
		s = ss[len(ss)-1]
	}
	// *T to T
	s = strings.TrimLeft(s, "*")
	if yes, _ := BuiltInTypes[s]; yes {
		return s, false
	}
	return s, true
}
