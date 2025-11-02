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

	type coverResponse struct {
		Files     []FileInfo
		Packages  []PackageInfo
		Functions []FunctionInfo
	}
	response := &coverResponse{}
	response.Files = ByFile(parsed)
	response.Packages = ByPackage(parsed)
	response.Functions = ByFunction(parsed)

	if cfg.GroupByFiles {
		return printCoverage[FileInfo](response.Files, encoder)
	}
	if cfg.GroupByPackage {
		return printCoverage[PackageInfo](response.Packages, encoder)
	}
	if cfg.GroupByFunction {
		return printCoverage[FunctionInfo](response.Functions, encoder)
	}

	encoder = json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func printCoverage[T fmt.Stringer](data []T, encoder *json.Encoder) error {
	if encoder != nil {
		return encoder.Encode(data)
	}
	for _, f := range data {
		fmt.Println(f.String())
	}
	return nil
}
