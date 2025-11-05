package report

import (
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
	"github.com/titpetric/exp/cmd/go-fsck/report/testusage"
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
	getDef := func(in *model.Definition) *model.Definition {
		for _, def := range defs {
			if def.Package.Equal(in.Package) {
				return def
			}
		}
		return nil
	}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, true, cfg.verbose)
		if err != nil {
			return nil, err
		}

		for _, in := range d {
			def := getDef(in)
			if def != nil {
				def.Merge(in)
				continue
			}
			defs = append(defs, in)
		}
	}

	return defs, nil
}

func report(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if len(defs) == 0 || err != nil {
		return fmt.Errorf("error getting definitions: %w, len %d", err, len(defs))
	}

	report, err := testusage.NewReport(defs)
	if err != nil {
		return fmt.Errorf("error generating report: %w", err)
	}

	fmt.Println(report)
	return nil
	// return json.NewEncoder(os.Stdout).Encode(report)
}
