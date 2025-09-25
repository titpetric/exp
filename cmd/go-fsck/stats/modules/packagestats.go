package modules

import (
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

type PackageStats struct {
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Functions int            `json:"functions"`
	Types     int            `json:"types"`
	Vars      int            `json:"vars"`
	Consts    int            `json:"consts"`
	Imports   map[string]int `json:"imports"` // import path → usage count
}

// NewPackageStats computes statistics per package for a Definition.
func NewPackageStats(def *model.Definition) []PackageStats {
	importCounts := make(map[string]int)

	for _, imp := range def.Imports {
		for _, v := range imp {
			importCounts[v]++
		}
	}

	stats := PackageStats{
		Name:      def.Package.Package,
		Path:      def.Path,
		Functions: len(def.Funcs),
		Types:     len(def.Types),
		Vars:      len(def.Vars),
		Consts:    len(def.Consts),
		Imports:   importCounts,
	}

	return []PackageStats{stats}
}
