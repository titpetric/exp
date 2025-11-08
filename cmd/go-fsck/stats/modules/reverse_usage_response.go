package modules

import (
	"fmt"
	"sort"
	"strings"
)

type ReverseUsageResponse struct {
	References map[string]map[string]int `json:"references"`
}

func NewReverseUsageResponse() ReverseUsageResponse {
	return ReverseUsageResponse{
		References: make(map[string]map[string]int),
	}
}

func (i ReverseUsageResponse) String() string {
	type kv struct {
		name string
		used int
	}
	var (
		mostUsed      string
		mostUsedCount int
		uniqueImports int = len(i.References)
		averageUse    float64
	)

	var totalUse int
	imports := []kv{}
	for pkg, usedRefs := range i.References {
		var used int
		for _, delta := range usedRefs {
			used += delta
		}

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
		message := fmt.Sprintf("- %s referenced %d times", p.name, p.used)
		importsList = append(importsList, message)
	}

	averageUse = float64(totalUse) / float64(uniqueImports)

	return fmt.Sprintf(
		"The package imports %d unique imports. Each import is used on average of %.2f times. The most used import is %s, with %d uses.\n\n"+
			"The most referenced imports:\n\n%s\n",
		uniqueImports, averageUse, mostUsed, mostUsedCount,
		strings.Join(importsList, "\n"),
	)
}

func (r *ReverseUsageResponse) Merge(in ReverseUsageResponse) {
	for k, v := range in.References {
		if _, ok := r.References[k]; !ok {
			r.References[k] = make(map[string]int, len(v))
		}
		for i, j := range v {
			r.References[k][i] += j
		}
	}
}
