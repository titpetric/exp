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

	parsed := Parse(lines)

	type coverResponse struct {
		Files     []FileInfo
		Packages  []PackageInfo
		Functions []FunctionInfo
	}
	response := &coverResponse{}
	response.Files = ByFile(parsed)
	response.Packages = ByPackage(parsed)
	response.Functions = ByFunction(parsed)

	if cfg.SkipUncovered {
		// Filter the aggregates rather than the functions feeding them: a
		// file's average must still count its uncovered functions, or a file
		// with one covered function out of twenty would report as covered.
		response.Files = covered(response.Files, func(r FileInfo) float64 { return r.Coverage })
		response.Packages = covered(response.Packages, func(r PackageInfo) float64 { return r.Coverage })
		response.Functions = covered(response.Functions, func(r FunctionInfo) float64 { return r.Coverage })
	}

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

// covered keeps the rows with any coverage, dropping the fully uncovered.
func covered[T any](rows []T, coverage func(T) float64) []T {
	var out []T
	for _, r := range rows {
		if coverage(r) > 0 {
			out = append(out, r)
		}
	}
	return out
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
