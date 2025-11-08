package modules

import (
	"fmt"
	"sort"
	"strings"
)

// ImportStatsResponse holds the number of references per imported package.
type ImportStatsResponse struct {
	// Holds a map of use from packages
	Imported map[string]int `json:"imported_from_packages"`
	// Holds a map of use from files
	ImportedFromFiles map[string]int `json:"imported_from_files"`
	// Holds a map of referenced imports
	Referenced map[string]int `json:"referenced"`
}

func (i ImportStatsResponse) String() string {
	type kv struct {
		name string
		used int
	}
	var (
		mostUsed      string
		mostUsedCount int
		uniqueImports int = len(i.Imported)
		averageUse    float64
	)

	var totalUse int
	imports := []kv{}
	for pkg, used := range i.Imported {
		imports = append(imports, kv{pkg, used})
		if used > mostUsedCount {
			mostUsedCount = used
			mostUsed = pkg
		}
		totalUse += used
	}

	sort.Slice(imports, func(i, j int) bool {
		if imports[i].used != imports[j].used {
			return imports[i].used > imports[j].used
		}
		return strings.Compare(imports[i].name, imports[j].name) > 0
	})

	if len(imports) > 10 {
		imports = imports[:10]
	}
	importsList := make([]string, 0, len(imports))
	for _, p := range imports {
		fileUse := i.ImportedFromFiles[p.name]
		refUse := i.Referenced[p.name]
		message := fmt.Sprintf("- %s, imported from %d packages, imported %d times, referenced %d times", p.name, p.used, fileUse, refUse)
		importsList = append(importsList, message)
	}

	averageUse = float64(totalUse) / float64(uniqueImports)

	return fmt.Sprintf(
		"The package imports %d unique imports. Each import is used on average of %.2f times. The most used import is %s, with %d uses.\n\n"+
			"The most used third party imports:\n\n%s\n",
		uniqueImports, averageUse, mostUsed, mostUsedCount,
		strings.Join(importsList, "\n"),
	)
}

func (i *ImportStatsResponse) Merge(in ImportStatsResponse) {
	for k, v := range in.Referenced {
		i.Referenced[k] += v
	}
	for k, v := range in.Imported {
		i.Imported[k] += v
	}
	for k, v := range in.ImportedFromFiles {
		i.ImportedFromFiles[k] += v
	}
}

func NewImportStatsResponse() ImportStatsResponse {
	return ImportStatsResponse{
		Referenced:        make(map[string]int),
		Imported:          make(map[string]int),
		ImportedFromFiles: make(map[string]int),
	}
}
