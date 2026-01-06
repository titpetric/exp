package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/covertrace/model"
)

func main() {
	var inputFile string
	var name string
	var skipSummary bool
	var outputJSON bool

	pflag.StringVarP(&inputFile, "input", "i", "", "Input coverage file")
	pflag.StringVarP(&name, "name", "n", "", "Test name")
	pflag.BoolVar(&skipSummary, "skip-summary", false, "Skip summary output")
	pflag.BoolVar(&outputJSON, "json", false, "Output detailed data in JSON format")
	pflag.Parse()

	if inputFile == "" {
		fmt.Println("Usage: go run main.go -i <coverage_file> [--skip-summary] [--json]")
		return
	}

	coverageData, err := model.ParseCoverageFile(inputFile)
	if err != nil {
		fmt.Println("Error parsing coverage file:", err)
		return
	}

	// Get all packages from the coverage file
	allPackages, err := model.ParseAllPackages(inputFile)
	if err != nil {
		fmt.Println("Error parsing packages:", err)
		return
	}

	structMap := make(map[string]map[string]int)
	structPackageMap := make(map[string]string) // map receiver -> packageName
	funcMap := make(map[string]int)
	packageMap := make(map[string]string)
	packageSet := make(map[string]bool)
	for _, p := range allPackages {
		packageSet[p] = true
	}

	for i := 0; i < len(coverageData); {
		cov := coverageData[i]

		symbol, receiver, coverage, err := getSymbolAndCoverage(cov.File, cov.StartLine, cov.EndLine, cov.NumStmts, cov.NumCov)
		if err != nil {
			fmt.Println("Error getting symbol or coverage:", err)
			return
		}

		if coverage > 0 {
			coverageData[i].Symbol = symbol
			coverageData[i].Receiver = receiver
			coverageData[i].Coverage = coverage

			// Get full package path from raw file
			fullPkgPath := getFullPackagePath(cov.PackageName, cov.RawFile)

			if receiver != "" {
				if structMap[receiver] == nil {
					structMap[receiver] = make(map[string]int)
				}
				structMap[receiver][symbol] += cov.NumCov
				structPackageMap[receiver] = fullPkgPath
				funcMap[fmt.Sprintf("%s.%s", receiver, symbol)] += cov.NumCov
			} else {
				funcMap[symbol] += cov.NumCov
				packageMap[symbol] = fullPkgPath
			}
			i++
		} else {
			coverageData = append(coverageData[:i], coverageData[i+1:]...)
		}
	}

	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}

	result := &model.Result{
		Name:     name,
		Packages: packages,
	}
	result.PopulateFromMaps(structMap, structPackageMap, funcMap, packageMap)

	if skipSummary {
		if outputJSON {
			printJSON(coverageData)
		} else {
			printRawYaml(coverageData)
		}
	} else {
		if outputJSON {
			printSummaryJSON(result)
		} else {
			summarizeYaml(result)
		}
	}
}

func getFullPackagePath(basePkg, rawFile string) string {
	// rawFile is like github.com/titpetric/atkins-ci/treeview/sorter.go
	// basePkg is like github.com/titpetric/atkins-ci
	// Extract the directory to get the full package path

	dir := rawFile
	lastSlash := strings.LastIndex(dir, "/")
	if lastSlash != -1 {
		dir = dir[:lastSlash]
	}

	return dir
}

func summarizeYaml(result *model.Result) {
	yamlData, err := yaml.Marshal(result)
	if err != nil {
		fmt.Println("Error marshalling to YAML:", err)
		return
	}
	fmt.Println(string(yamlData))
}

func printJSON(coverageData []model.CoverageInfo) {
	jsonOutput, err := json.MarshalIndent(coverageData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	fmt.Println(string(jsonOutput))
}

func printRawYaml(coverageData []model.CoverageInfo) {
	yamlOutput, err := yaml.Marshal(coverageData)
	if err != nil {
		fmt.Println("Error marshalling to YAML:", err)
		return
	}
	fmt.Println(string(yamlOutput))
}

func printSummaryJSON(result *model.Result) {
	summaryJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling summary to JSON:", err)
		return
	}
	fmt.Println(string(summaryJSON))
}
