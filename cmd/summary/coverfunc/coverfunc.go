package coverfunc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/titpetric/exp/cmd/summary/internal"
)

func coverfunc(cfg *options) error {
	lines, err := internal.ReadFields(os.Stdin)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if cfg == nil {
		fmt.Println("Options not provided. Rendering data.")
		return coverPackages(nil, lines, encoder)
	}

	if cfg.RenderJSON {
		if cfg.GroupByFiles {
			return coverFiles(cfg, lines, encoder)
		}
		return coverPackages(cfg, lines, encoder)
	}

	if cfg.GroupByFiles {
		return coverFiles(cfg, lines, nil)
	}
	return coverPackages(cfg, lines, nil)
}

func coverFiles(cfg *options, lines [][]string, encoder *json.Encoder) error {
	parsed := Parse(lines, cfg.SkipUncovered)
	files := ByFile(parsed)
	if encoder != nil {
		return encoder.Encode(files)
	}
	for _, f := range files {
		fmt.Println(f.String())
	}
	return nil
}

func coverPackages(cfg *options, lines [][]string, encoder *json.Encoder) error {
	parsed := Parse(lines, cfg.SkipUncovered)
	pkgs := ByPackage(parsed)
	if encoder != nil {
		return encoder.Encode(pkgs)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.String())
	}
	return nil
}
