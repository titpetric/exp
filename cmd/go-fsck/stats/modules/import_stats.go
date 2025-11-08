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

	seen := map[string]bool{}

	pkg := def.Package.ImportPath
	if !strings.Contains(pkg, ".") {
		return result
	}

	// exclude known prefix
	if strings.HasPrefix(pkg, "github.com/dolthub/dolt/go") {
		return result
	}

	// a file has multiple declarations
	for _, d := range decls {
		imports := def.Imports.Get(d.File)
		importMap, _ := def.Imports.Map(imports)

		for _, d := range decls {
			imports := def.Imports.Get(d.File)
			importMap, _ := def.Imports.Map(imports)

			if _, ok := seen[d.File]; !ok {
				result.Imported[pkg] = 1
				for _, pkg := range importMap {
					result.ImportedFromFiles[pkg]++
				}
				seen[d.File] = true
			}
		}

		for short, symbols := range d.References {
			pkg, ok := importMap[short]
			if !ok {
				pkg = short
			}

			// don't count standard library use
			if !strings.Contains(pkg, ".") {
				continue
			}

			// exclude known prefix
			if strings.HasPrefix(pkg, "github.com/dolthub/dolt/go") {
				continue
			}

			result.Referenced[pkg] += len(symbols)
		}
	}

	return result
}
