package model

// DefinitionList holds a list of Go packages.
type DefinitionList []*Definition

func (p DefinitionList) Walk(matchfn func(d *Definition)) {
	for _, decl := range p {
		matchfn(decl)
	}
}

func (p DefinitionList) Find(matchfn func(d *Definition) bool) *Definition {
	for _, decl := range p {
		if matchfn(decl) {
			return decl
		}
	}
	return nil
}

func (p DefinitionList) Filter(matchfn func(d *Definition) bool) (result []*Definition) {
	for _, decl := range p {
		if matchfn(decl) {
			result = append(result, decl)
		}
	}
	return
}
