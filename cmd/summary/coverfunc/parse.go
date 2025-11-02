package coverfunc

import (
	"path"
	"strconv"
	"strings"
)

// Parse parses the coverage data into CoverageInfo.
func Parse(data [][]string, skipUncovered bool) []CoverageInfo {
	var coverageInfos []CoverageInfo

	for _, line := range data {
		filenameAndLine := strings.Split(line[0], ":")
		filename := filenameAndLine[0]
		if filename == "total" {
			continue
		}

		// Assuming that line number is always present and can be parsed
		lineNumber, _ := strconv.Atoi(filenameAndLine[1])

		info := CoverageInfo{
			File:     path.Base(filename),
			Filename: filename,
			Package:  path.Dir(filename),
			Line:     lineNumber,
			Function: line[1],
		}
		percent, _ := strconv.ParseFloat(strings.TrimSuffix(line[2], "%"), 64)
		info.Coverage = percent

		if percent <= 0 && skipUncovered {
			continue
		}

		coverageInfos = append(coverageInfos, info)
	}

	return coverageInfos
}
