package model

import (
	"go/ast"
	"sort"
	"strings"
)

// DeclarationList holds a list of Go symbols.
type DeclarationList []*Declaration

func (p *DeclarationList) Append(in ...*Declaration) {
	*p = append(*p, in...)
}

func (p *DeclarationList) AppendUnique(in ...*Declaration) {
	for _, i := range in {
		shouldAppend := true
		for _, decl := range *p {
			if decl.Equal(i) {
				shouldAppend = false
				break
			}
		}

		if shouldAppend {
			*p = append(*p, i)
		}
	}
	p.Sort()
}

func (p *DeclarationList) ClearNonTestFiles() {
	result := DeclarationList{}
	for _, decl := range *p {
		if !strings.HasSuffix(decl.File, "_test.go") {
			continue
		}
		result.Append(decl)
	}
	*p = result
}

func (p *DeclarationList) ClearSource() {
	for _, decl := range *p {
		decl.Source = ""
	}
}

func (p *DeclarationList) ClearTestFiles() {
	result := DeclarationList{}
	for _, decl := range *p {
		if strings.HasSuffix(decl.File, "_test.go") {
			continue
		}
		result.Append(decl)
	}
	*p = result
}

func (p *DeclarationList) Sort() {
	sort.Slice(*p, func(i, j int) bool {
		a, b := (*p)[i], (*p)[j]
		if a.Kind != b.Kind {
			indexOf := map[DeclarationKind]int{
				CommentKind: 0,
				ImportKind:  1,
				ConstKind:   2,
				StructKind:  3,
				TypeKind:    4,
				VarKind:     5,
				FuncKind:    6,
			}
			return indexOf[a.Kind] < indexOf[b.Kind]
		}
		ae, be := ast.IsExported(a.Name), ast.IsExported(b.Name)
		if ae != be {
			return ae
		}

		if a.Receiver != b.Receiver {
			if a.Receiver == "" {
				return true
			}
			return a.Receiver < b.Receiver
		}

		if a.Signature != b.Signature {
			return a.Signature < b.Signature
		}

		return a.Name < b.Name
	})
}

func (p DeclarationList) Exported() (result []*Declaration) {
	for _, decl := range p {
		if decl.IsExported() {
			result = append(result, decl)
		}
	}
	return
}

func (p DeclarationList) Filter(matchfn func(d *Declaration) bool) (result []*Declaration) {
	for _, decl := range p {
		if matchfn(decl) {
			result = append(result, decl)
		}
	}
	return
}

func (p DeclarationList) Find(matchfn func(d *Declaration) bool) *Declaration {
	for _, decl := range p {
		if matchfn(decl) {
			return decl
		}
	}
	return nil
}

func (p DeclarationList) Walk(matchfn func(d *Declaration)) {
	for _, decl := range p {
		matchfn(decl)
	}
}
