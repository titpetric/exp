package model

import (
	"fmt"
)

type Ref struct {
	Package  *Package
	Receiver string
	Name     string
}

func (r Ref) String() string {
	if r.Receiver != "" {
		return fmt.Sprintf("%s.%s.%s", r.Package.Name(), TypeRef(r.Receiver), TypeRef(r.Name))
	}
	return fmt.Sprintf("%s.%s", r.Package.Name(), TypeRef(r.Name))
}
