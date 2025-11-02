package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// Read the exported go-fsck.json data.
	defs, err := loader.ReadFile(cfg.inputFile)
	if err == nil {
		return defs, nil
	}

	// list current local packages
	packages, err := internal.ListPackages(".", ".")
	if err != nil {
		return nil, err
	}

	defs = []*model.Definition{}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, cfg.verbose)
		if err != nil {
			return nil, err
		}

		defs = append(defs, d...)
	}

	return defs, nil
}

var (
	coveragePass = fmt.Sprintf("%c", '\U00002705')
	coverageFail = fmt.Sprintf("%c", '\U0000274C')
)

type CoverageInfo struct {
	File, Package, Function string
	Coverage                float64
	Cognitive               int
}

func loadCoverage(name string) ([]CoverageInfo, error) {
	var result []CoverageInfo

	// We can just print coverage.
	if name == "" {
		return result, nil
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s: %w", name, err)
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("Error decoding %s: %w", name, err)
	}

	return result, err
}

func coverage(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	coverinfo, err := loadCoverage(cfg.coverageFile)
	if err != nil {
		return err
	}

	findPackage := func(defs []*model.Definition, name string) *model.Definition {
		for _, def := range defs {
			if name == def.Package.ImportPath {
				return def
			}
		}
		return nil
	}

	var total int
	for _, info := range coverinfo {
		p := findPackage(defs, info.Package)
		if p == nil {
			return fmt.Errorf("Can't find package by name: %s", info.Package)
		}

		f := p.Funcs.Find(func(d *model.Declaration) bool {
			if d.Kind != model.FuncKind {
				return false
			}
			return d.Name == info.Function && d.File == info.File
		})
		if f == nil {
			return fmt.Errorf("Can't find function by name: %s, package: %s", info.Function, info.Package)
		}

		if f.Complexity != nil {
			f.Complexity.Coverage = info.Coverage
			if info.Coverage > 0 {
				total++
			}
		}
	}

	if cfg.outputFile != "" {
		b, err := json.MarshalIndent(defs, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(cfg.outputFile, b, 0644); err != nil {
			return err
		}

		fmt.Printf("Wrote coverage information for %d/%d functions to %s\n", total, len(coverinfo), cfg.outputFile)
	} else {
		var result []CoverageInfo
		for _, def := range defs {
			fns := def.Funcs.Filter(func(d *model.Declaration) bool {
				if cfg.verbose {
					return true
				}
				return d.Complexity != nil && d.Complexity.Coverage > 0
			})
			for _, fn := range fns {
				info := CoverageInfo{
					Package:   def.Package.ImportPath,
					Function:  combined(fn.Receiver, fn.Name),
					Coverage:  fn.Complexity.Coverage,
					Cognitive: fn.Complexity.Cognitive,
				}
				result = append(result, info)
			}
		}

		sort.Slice(result, func(i, j int) bool {
			var k, v = result[i], result[j]
			if k.Package != v.Package {
				return strings.Compare(k.Package, v.Package) < 0
			}
			if k.Function != v.Function {
				return strings.Compare(k.Function, v.Function) < 0
			}
			return k.Coverage > v.Coverage
		})

		if cfg.json {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(result)
		}

		// Encode aggregated results as markdown.
		data := [][]string{}
		for idx, r := range result {
			pass := r.Cognitive == 0 || r.Coverage > 0
			passText := coveragePass
			if !pass {
				passText = coverageFail
			}

			data = append(data, []string{fmt.Sprint(idx), passText, r.Package, r.Function, fmt.Sprintf("%.2f%%", r.Coverage), fmt.Sprint(r.Cognitive)})
		}

		table, err := markdown.NewTableFormatterBuilder().WithPrettyPrint().Build("#", "Status", "Package", "Function", "Coverage", "Cognit").Format(data)
		if err != nil {
			return err
		}

		fmt.Println(table)
	}

	return nil
}

func combined(receiver, name string) string {
	if receiver != "" {
		return strings.TrimLeft(receiver, "*") + "." + name
	}
	return name
}
