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

func DBTag(tag string) string {
	return reflect.StructTag(tag).Get("db")
}

func DBTagName(tag string) string {
	out := DBTag(tag)
	parts := strings.SplitN(out, ",", 2)
	return parts[0]
}
