package coverfunc

import (
	"path"
	"sort"
	"strings"
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
		packageMap[packageName] = append(packageMap[packageName], info.Percent)
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
		if packageInfos[i].Coverage != packageInfos[j].Coverage {
			return packageInfos[i].Coverage > packageInfos[j].Coverage
		}
		return packageInfos[i].Package > packageInfos[j].Package
	})

	return cleanPackageNames(packageInfos)
}

// cleanPackageNames removes the common prefix of the package name, shortening to a relative path.
func cleanPackageNames(packageInfos []PackageInfo) []PackageInfo {
	if len(packageInfos) == 0 {
		return nil
	}

	// Find the common prefix among package names
	commonPrefix := findCommonPrefix(packageInfos)

	// Shorten package names by removing the common prefix
	var cleanedPackageInfos []PackageInfo
	for _, pkg := range packageInfos {
		relativePath := strings.TrimPrefix(pkg.Package, commonPrefix)
		relativePath = strings.TrimPrefix(relativePath, "/")
		cleanedPackageInfos = append(cleanedPackageInfos, PackageInfo{
			Package:   relativePath,
			Functions: pkg.Functions,
			Coverage:  pkg.Coverage,
		})
	}

	return cleanedPackageInfos
}

// findCommonPrefix finds the common prefix among package names.
func findCommonPrefix(packageInfos []PackageInfo) string {
	if len(packageInfos) == 0 {
		return ""
	}

	prefix := packageInfos[0].Package
	for _, pkg := range packageInfos[1:] {
		for !strings.HasPrefix(pkg.Package, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}

	// Trim to ensure a clean relative path
	for len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix = prefix[:len(prefix)-1]
	}

	return prefix
}
