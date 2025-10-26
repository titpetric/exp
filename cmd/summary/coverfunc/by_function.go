package coverfunc

import "sort"

// ByFunction basically just prints per function coverage information.
// This requires no grouping, just a conversion.
func ByFunction(coverageInfos []CoverageInfo) []FunctionInfo {
	var result []FunctionInfo

	for _, info := range coverageInfos {
		result = append(result, FunctionInfo{
			Package:  info.GetPackage(),
			Function: info.Function,
			Coverage: info.Percent,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Coverage > result[j].Coverage
	})

	return result
}
