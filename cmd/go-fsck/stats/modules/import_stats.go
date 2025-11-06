package modules

import (
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ImportStats produces a grouping of short package names imported.
func ImportStats(defs model.DefinitionList) ImportStatsResponse {
	res := NewImportStatsResponse()
	for _, def := range defs {
		res.Merge(ImportStatsForDefinition(def))
	}
	return res
}

// ImportStatsForDefinition computes the number of times each imported package is referenced.
func ImportStatsForDefinition(def *model.Definition) ImportStatsResponse {
	result := NewImportStatsResponse()
	decls := def.DeclarationList()

	for _, d := range decls {
		imports := def.Imports.Get(d.File)
		importMap, _ := def.Imports.Map(imports)

		for pkg, refs := range d.References {
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

				result.Imported[pkgName] += len(refs)
			} else {
				result.Imported[pkg] += len(refs)
			}
		}
	}

	return result
}
