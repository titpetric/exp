package coverfunc

import (
	"path"
	"sort"
)

// ByPackage summarizes coverage info by package.
func ByPackage(coverageInfos []CoverageInfo) []PackageInfo {
	packageMap := make(map[string][]float64)
	functionsMap := make(map[string]int)

	for _, info := range coverageInfos {
		packageName := path.Dir(info.Filename)
		if _, ok := packageMap[packageName]; !ok {
			packageMap[packageName] = []float64{}
		}
		packageMap[packageName] = append(packageMap[packageName], info.Coverage)
		functionsMap[packageName]++
	}

	var packageInfos []PackageInfo

	for packageName, percentages := range packageMap {
		sum := 0.0
		for _, percent := range percentages {
			sum += percent
		}
		avgCoverage := sum / float64(len(percentages))
		packageInfos = append(packageInfos, PackageInfo{
			Package:   packageName,
			Functions: functionsMap[packageName],
			Coverage:  avgCoverage,
		})
	}

	sort.Slice(packageInfos, func(i, j int) bool {
		if packageInfos[i].Package != packageInfos[j].Package {
			return packageInfos[i].Package < packageInfos[j].Package
		}
		if packageInfos[i].Coverage != packageInfos[j].Coverage {
			return packageInfos[i].Coverage > packageInfos[j].Coverage
		}
		return false
	})

	return packageInfos
}
