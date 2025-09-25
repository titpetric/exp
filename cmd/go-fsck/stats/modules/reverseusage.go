package modules

import (
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ReverseUsage tracks how many declarations use each symbol in each imported package
type ReverseUsage struct {
	Results map[string]map[string]int `json:"results"`
}

// NewReverseUsage computes reverse usage counts for all package symbols
func NewReverseUsage(def *model.Definition) ReverseUsage {
	result := make(map[string]map[string]int)

	allDecls := model.DeclarationList{}
	allDecls.Append(def.Funcs...)
	allDecls.Append(def.Types...)
	allDecls.Append(def.Vars...)
	allDecls.Append(def.Consts...)

	for _, d := range allDecls {
		for pkg, symbols := range d.References {
			if _, ok := result[pkg]; !ok {
				result[pkg] = make(map[string]int)
			}
			for _, sym := range symbols {
				result[pkg][sym]++
			}
		}
	}

	return ReverseUsage{Results: result}
}
