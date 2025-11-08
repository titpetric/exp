package model

import (
	"go/ast"
	"strings"
)

// Declaration holds information about a go symbol.
type Declaration struct {
	Kind DeclarationKind
	Type string `json:",omitempty"`

	File string
	Line int `json:",omitempty"`

	SelfContained bool

	// This is not encoded to json, it's computed on load.
	Imports []string `json:"-"`

	References StringSet `json:",omitempty"`
	Globals    StringSet `json:",omitempty"`

	Doc string `json:",omitempty"`

	Name     string   `json:",omitempty"`
	Names    []string `json:",omitempty"`
	Receiver string   `json:",omitempty"`

	Fields FieldList `json:",omitempty"`

	Arguments []string `json:",omitempty"`
	Returns   []string `json:",omitempty"`

	Signature string `json:",omitempty"`
	Source    string `json:",omitempty"`

	Complexity *Complexity `json:",omitempty"`
}

func (f *Declaration) TypeRef() string {
	return TypeRef(f.Type)
}

func (f *Declaration) ReceiverTypeRef() string {
	return TypeRef(f.Receiver)
}

func (d *Declaration) HasName(find string) bool {
	for _, name := range d.Names {
		if name == find {
			return true
		}
	}
	return d.Name == find
}

func (d *Declaration) IsExported() bool {
	if d.Receiver != "" && !ast.IsExported(TypeRef(d.Receiver)) {
		return false
	}

	for _, name := range d.Names {
		if ast.IsExported(name) {
			return true
		}
	}
	return ast.IsExported(d.Name)
}

func (d *Declaration) Equal(in *Declaration) bool {
	if d.File == in.File && d.Kind == in.Kind && d.Name == in.Name {
		return true
	}
	return false
}

func (d *Declaration) Keys() []string {
	trimPath := "*."
	if d.Name != "" {
		return []string{
			strings.Trim(d.Receiver+"."+d.Name, trimPath),
		}
	}
	if len(d.Names) != 0 {
		result := make([]string, len(d.Names))
		for k, v := range d.Names {
			result[k] = strings.Trim(d.Receiver+"."+v, trimPath)
		}
		return result
	}
	return nil
}
