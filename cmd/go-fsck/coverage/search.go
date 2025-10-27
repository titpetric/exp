package coverage

import (
	"encoding/json"
	"fmt"
	"os"

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

type CoverageInfo struct {
	Package, Function string
	Coverage          float64
}

func loadCoverage(name string) ([]CoverageInfo, error) {
	var result []CoverageInfo

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

	for _, info := range coverinfo {
		p := findPackage(defs, info.Package)
		if p == nil {
			return fmt.Errorf("Can't find package by name: %s", info.Package)
		}

		f := p.Funcs.Find(func(d *model.Declaration) bool {
			if d.Kind != model.FuncKind {
				return false
			}
			return d.Name == info.Function
		})
		if f == nil {
			return fmt.Errorf("Can't find function by name: %s, package: %s", info.Function, info.Package)
		}

		if f.Complexity != nil {
			f.Complexity.Coverage = info.Coverage
		}
	}

	b, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfg.outputFile, b, 0644); err != nil {
		return err
	}

	fmt.Printf("Wrote coverage information for %d functions to %s\n", len(coverinfo), cfg.outputFile)

	return nil
}
