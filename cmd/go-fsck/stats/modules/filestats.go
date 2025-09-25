package modules

import (
	"path"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// FileStats holds statistics for a single file, including imports
type FileStats struct {
	Name      string         `json:"name"`
	Functions int            `json:"functions"`
	Types     int            `json:"types"`
	Vars      int            `json:"vars"`
	Consts    int            `json:"consts"`
	Imports   map[string]int `json:"imports"` // import path → usage count
}

// NewFileStats computes statistics per file across all symbols in a Definition.
func NewFileStats(def *model.Definition) []FileStats {
	statsMap := make(map[string]*FileStats)

	countDecls := func(decls model.DeclarationList, kind string) {
		for _, d := range decls {
			name := path.Join(def.Path, d.File)
			if _, ok := statsMap[name]; !ok {
				statsMap[name] = &FileStats{
					Name:    name,
					Imports: make(map[string]int),
				}
			}

			imports, ok := def.Imports[d.File]
			if ok {
				for _, imp := range imports {
					statsMap[name].Imports[imp] = 1
				}
			}

			switch kind {
			case "func":
				statsMap[name].Functions++
			case "type":
				statsMap[name].Types++
			case "var":
				statsMap[name].Vars++
			case "const":
				statsMap[name].Consts++
			}

		}
	}

	countDecls(def.Funcs, "func")
	countDecls(def.Types, "type")
	countDecls(def.Vars, "var")
	countDecls(def.Consts, "const")

	results := make([]FileStats, 0, len(statsMap))
	for _, v := range statsMap {
		results = append(results, *v)
	}

	return results
}
