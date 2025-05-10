package ast

import (
	"reflect"
	"strings"
)

func JSONTag(tag string) string {
	return reflect.StructTag(tag).Get("json")
}

func JSONTagName(tag string) string {
	out := JSONTag(tag)
	parts := strings.SplitN(out, ",", 2)
	return parts[0]
}
