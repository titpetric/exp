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
		d, err := loader.Load(pkg, true, cfg.verbose)
		if err != nil {
			return nil, err
		}

		defs = append(defs, d...)
	}

	return defs, nil
}

func report(title string, value any) {
	fmt.Println("##", title)
	fmt.Println()
	if v, ok := value.(fmt.Stringer); ok {
		fmt.Println(v.String())
	} else {
		j, _ := json.MarshalIndent(value, "", "  ")
		fmt.Println(string(j))
	}
	fmt.Println()
}

func stats(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	report("Documentation", modules.Documentation(defs))
	report("Package stats", modules.PackageStats(defs))
	report("Import usage", modules.ImportStats(defs))

	for _, def := range defs {
		fmt.Println("#", def.ImportPath)
		fmt.Println()
		report("File stats", modules.NewFileStats(def))
		report("Reverse usage", modules.NewReverseUsage(def))
	}

	return nil
}
