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

	var encoder *json.Encoder

	if cfg.RenderJSON {
		encoder = json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
	}

	parsed := Parse(lines, cfg.SkipUncovered)

	if cfg.GroupByFiles {
		return coverFiles(parsed, encoder)
	}
	if cfg.GroupByPackage {
		return coverPackages(parsed, encoder)
	}
	return coverFunctions(parsed, encoder)
}

func coverFunctions(parsed []CoverageInfo, encoder *json.Encoder) error {
	files := ByFunction(parsed)
	if encoder != nil {
		return encoder.Encode(files)
	}
	for _, f := range files {
		fmt.Println(f.String())
	}
	return nil
}

func coverFiles(parsed []CoverageInfo, encoder *json.Encoder) error {
	files := ByFile(parsed)
	if encoder != nil {
		return encoder.Encode(files)
	}
	for _, f := range files {
		fmt.Println(f.String())
	}
	return nil
}

func coverPackages(parsed []CoverageInfo, encoder *json.Encoder) error {
	pkgs := ByPackage(parsed)
	if encoder != nil {
		return encoder.Encode(pkgs)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.String())
	}
	return nil
}
