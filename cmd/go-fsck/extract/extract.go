package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// list current local packages
	var pattern string
	if pattern = "."; cfg.recursive {
		pattern = "./..."
	}

	packages, err := internal.ListPackages(cfg.sourcePath, pattern)
	if err != nil {
		return nil, err
	}

	defs := []*model.Definition{}

	for _, pkg := range packages {
		if !cfg.includeTests && pkg.TestPackage {
			continue
		}

		d, err := loader.Load(pkg, cfg.verbose)
		if err != nil {
			return nil, err
		}

		// White box test include whole package scope. Lie.
		if pkg.TestPackage {
			if !strings.HasSuffix(pkg.Package, "_test") {
				pkg.Package += "_test"
				pkg.ImportPath += "_test" // More about the binary, it's test scope even if not black box.
			}
		}

		for _, v := range d {
			v.Package.ID = pkg.ID
			v.Package.ImportPath = pkg.ImportPath
			v.Package.Path = pkg.Path
			v.Package.Package = pkg.Package
			v.Package.TestPackage = pkg.TestPackage
		}

		defs = append(defs, d...)
	}

	defs = unique(defs)

	for _, def := range defs {
		if !cfg.includeSources {
			def.ClearSource()
		}
		if !def.TestPackage || !cfg.includeTests {
			def.ClearTestFiles()
		}
		if def.TestPackage {
			def.ClearNonTestFiles()
		}
	}

	return defs, nil
}

func extract(cfg *options) error {
	definitions, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	output := os.Stdout
	switch cfg.outputFile {
	case "", "-":
	default:
		fmt.Println(cfg.outputFile)
		var err error
		output, err = os.Create(cfg.outputFile)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(output)
	if cfg.prettyJSON {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(definitions)
}
