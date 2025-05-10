package model

import (
	"strings"
)

// TypeRef aims to trim a type name to a reference type.
func TypeRef(name string) string {
	// trim variadic arg
	if strings.HasPrefix(name, "...") {
		name = name[3:]
	}

	// slice, array and map value
	// in terms of nesting, this is a hack
	if strings.HasPrefix(name, "[") || strings.HasPrefix(name, "map[") {
		name = strings.SplitN(name, "]", 2)[1]
	}

	// deref pointers
	return strings.TrimPrefix(name, "*")
}
