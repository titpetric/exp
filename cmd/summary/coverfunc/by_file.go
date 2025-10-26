package coverfunc

import "sort"

// ByFile summarizes coverage info by file.
func ByFile(coverageInfos []CoverageInfo) []FileInfo {
	fileMap := make(map[string][]float64)
	functionsMap := make(map[string]int)

	for _, info := range coverageInfos {
		if _, ok := fileMap[info.Filename]; !ok {
			fileMap[info.Filename] = []float64{}
		}
		fileMap[info.Filename] = append(fileMap[info.Filename], info.Percent)
		functionsMap[info.Filename]++
	}

	var fileInfos []FileInfo

	for filename, percentages := range fileMap {
		sum := 0.0
		for _, percent := range percentages {
			sum += percent
		}
		avgCoverage := sum / float64(len(percentages))
		fileInfos = append(fileInfos, FileInfo{
			Filename:  filename,
			Functions: functionsMap[filename],
			Coverage:  avgCoverage,
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Coverage > fileInfos[j].Coverage
	})

	return fileInfos
}
