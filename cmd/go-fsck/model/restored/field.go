package model

// Field holds details about a field definition.
type Field struct {
	// Name is the name of the field.
	Name string

	// Type is the literal type of the Go field.
	Type string `json:",omitempty"`

	// Path is the go path of this field starting from root object.
	Path string

	// Doc holds the field doc.
	Doc string `json:",omitempty"`

	// Comment holds the field comment text.
	Comment string `json:",omitempty"`

	// Tag is the go tag, unmodified.
	Tag string `json:",omitempty"`

	// JSONName is the corresponding json name of the field.
	// It's cleared if it's set to `-` (unexported).
	JSONName string

	// MapKey is the map key type, if this field is a map.
	MapKey string `json:",omitempty"`
}

func (f *Field) TypeRef() string {
	return TypeRef(f.Type)
}
