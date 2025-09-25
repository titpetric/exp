package stats

import (
	"encoding/json"
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
	"github.com/titpetric/exp/cmd/go-fsck/stats/modules"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// Read the exported go-fsck.json data.
	defs, err := loader.ReadFile(cfg.inputFile)
	if err == nil {
		return defs, nil
	}

	// list current local package
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

func report(title string, value any) {
	j, _ := json.MarshalIndent(value, "", "  ")

	fmt.Println(title)
	fmt.Println()
	fmt.Println(string(j))
	fmt.Println()
}

func stats(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	for _, def := range defs {
		report("File stats", modules.NewFileStats(def))
		report("Package usage", modules.NewPackagePollution(def))
		report("Package stats", modules.NewPackageStats(def))
		report("Reverse usage", modules.NewReverseUsage(def))
	}

	return nil
}
