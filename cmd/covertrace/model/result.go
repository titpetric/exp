package model

import "strings"

type Result struct {
	Name     string       `yaml:"name,omitempty"`
	Packages []string     `yaml:"packages"`
	Types    []StructType `yaml:"types"`
	Funcs    []GlobalFunc `yaml:"funcs"`
	Totals   TotalOutput  `yaml:"total"`
}

type StructType struct {
	Name        string        `yaml:"name"`
	PackageName string        `yaml:"packageName"`
	Funcs       []FuncDetails `yaml:"funcs"`
}

type FuncDetails struct {
	Name     string `yaml:"name"`
	Coverage int    `yaml:"coverage"`
}

type GlobalFunc struct {
	Name        string `yaml:"name"`
	PackageName string `yaml:"packageName"`
	Coverage    int    `yaml:"coverage"`
}

type TotalOutput struct {
	Coverage struct {
		Funcs   int `yaml:"funcs"`
		Structs int `yaml:"structs"`
		Total   int `yaml:"total"`
	} `yaml:"coverage"`
}

// PopulateFromMaps populates the Result with data from the working maps
func (r *Result) PopulateFromMaps(
	structMap map[string]map[string]int,
	structPackageMap map[string]string,
	funcMap map[string]int,
	packageMap map[string]string,
) {
	var structsCoverage int
	var globalsCoverage int

	for structName, funcs := range structMap {
		var functions []FuncDetails
		for funcName, covStmts := range funcs {
			functions = append(functions, FuncDetails{
				Name:     funcName,
				Coverage: covStmts,
			})
			structsCoverage += covStmts
		}
		r.Types = append(r.Types, StructType{
			Name:        structName,
			PackageName: structPackageMap[structName],
			Funcs:       functions,
		})
	}

	for funcName, covStmts := range funcMap {
		if !strings.Contains(funcName, ".") {
			r.Funcs = append(r.Funcs, GlobalFunc{
				Name:        funcName,
				PackageName: packageMap[funcName],
				Coverage:    covStmts,
			})
			globalsCoverage += covStmts
		}
	}

	totalCoverage := structsCoverage + globalsCoverage
	r.Totals = TotalOutput{
		Coverage: struct {
			Funcs   int `yaml:"funcs"`
			Structs int `yaml:"structs"`
			Total   int `yaml:"total"`
		}{
			Funcs:   globalsCoverage,
			Structs: structsCoverage,
			Total:   totalCoverage,
		},
	}
}
