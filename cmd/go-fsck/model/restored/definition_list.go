package model

// DefinitionList holds a list of Go packages.
type DefinitionList []*Definition

// Map over a DefinitionList
func Map[T any](defs DefinitionList, fn func(*Definition) T) []T {
	out := make([]T, 0, len(defs))
	for _, d := range defs {
		out = append(out, fn(d))
	}
	return out
}

// Reduce over a DefinitionList
func Reduce[T any, R any](defs DefinitionList, acc R, fn func(R, *Definition) R) R {
	for _, d := range defs {
		acc = fn(acc, d)
	}
	return acc
}

func (p DefinitionList) Filter(matchfn func(d *Definition) bool) (result []*Definition) {
	for _, decl := range p {
		if matchfn(decl) {
			result = append(result, decl)
		}
	}
	return
}

func (p DefinitionList) Find(matchfn func(d *Definition) bool) *Definition {
	for _, decl := range p {
		if matchfn(decl) {
			return decl
		}
	}
	return nil
}

func (p DefinitionList) Walk(matchfn func(d *Definition)) {
	for _, decl := range p {
		matchfn(decl)
	}
}
