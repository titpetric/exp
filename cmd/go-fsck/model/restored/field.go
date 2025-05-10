package model

import (
	"strings"
)

// Field holds details about a field definition.
type Field struct {
	// Name is the name of the field.
	Name string `json:"name"`

	// Type is the literal type of the Go field.
	Type string `json:"type"`

	// Path is the go path of this field starting from root object.
	Path string `json:"path"`

	// Doc holds the field doc.
	Doc string `json:"doc,omitempty"`

	// Comment holds the field comment text.
	Comment string `json:"comment,omitempty"`

	// Tag is the go tag, unmodified.
	Tag string `json:"tag,omitempty"`

	// JSONName is the corresponding json name of the field.
	// It's cleared if it's set to `-` (unexported).
	JSONName string `json:"json_name"`

	// MapKey is the map key type, if this field is a map.
	MapKey string `json:"map_key,omitempty"`
}

func (f *Field) TypeRef() string {
	return strings.TrimLeft(f.Type, "[]*")
}
