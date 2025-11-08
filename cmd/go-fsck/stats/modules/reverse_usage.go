package modules

import (
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func ReverseUsage(defs model.DefinitionList) ReverseUsageResponse {
	res := NewReverseUsageResponse()
	for _, def := range defs {
		res.Merge(ReverseUsageForDefinition(def))
	}
	return res
}

func ReverseUsageForDefinition(def *model.Definition) ReverseUsageResponse {
	response := NewReverseUsageResponse()
	decls := def.DeclarationList()

	for _, d := range decls {
		imports := def.Imports.Get(d.File)
		importMap, _ := def.Imports.Map(imports)

		for pkg, symbols := range d.References {
			pkgName, ok := importMap[pkg]
			if ok {
				// don't count standard library use
				if !strings.Contains(pkgName, ".") {
					continue
				}
				// exclude known prefix
				if strings.HasPrefix(pkgName, "github.com/dolthub/dolt/go") {
					continue
				}
			} else {
				pkgName = pkg
			}

			if _, ok := response.References[pkgName]; !ok {
				response.References[pkgName] = make(map[string]int)
			}
			for _, sym := range symbols {
				response.References[pkgName][sym]++
			}
		}
	}

	return response
}
