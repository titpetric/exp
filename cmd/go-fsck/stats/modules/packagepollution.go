package modules

import (
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// PackagePollution holds the number of references per imported package
type PackagePollution struct {
	Results map[string]int `json:"results"`
}

// NewPackagePollution computes the number of times each imported package is referenced
func NewPackagePollution(def *model.Definition) PackagePollution {
	result := make(map[string]int)

	allDecls := model.DeclarationList{}
	allDecls.Append(def.Funcs...)
	allDecls.Append(def.Types...)
	allDecls.Append(def.Vars...)
	allDecls.Append(def.Consts...)

	for _, d := range allDecls {
		for pkg, refs := range d.References {
			result[pkg] += len(refs)
		}
	}

	return PackagePollution{Results: result}
}
