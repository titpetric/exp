package model

// DeclarationKind is an enum of go symbol types.
type DeclarationKind string

func (d DeclarationKind) String() string {
	return string(d)
}
