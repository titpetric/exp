package modules

import (
	"slices"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func PackageStats(defs model.DefinitionList) PackageStatsResponse {
	res := NewPackageStatsResponse()
	for _, def := range defs {
		res.Merge(PackageStatsForDefinition(def))
	}
	return res
}

func PackageStatsForDefinition(def *model.Definition) PackageStatsResponse {
	res := NewPackageStatsResponse()
	res.Functions = len(def.Funcs)
	res.Types = len(def.Types)
	res.Vars = len(def.Vars)
	res.Consts = len(def.Consts)

	for _, imp := range def.Imports {
		for _, v := range imp {
			res.Imports[v]++
		}
	}

	files := []string{}
	decls := def.DeclarationList()
	for _, decl := range decls {
		if slices.Contains(files, decl.File) {
			continue
		}
		files = append(files, decl.File)
	}
	res.Files = len(files)

	return res
}
