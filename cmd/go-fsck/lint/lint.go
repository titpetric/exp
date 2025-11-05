package lint

import (
	"errors"
	"fmt"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// list current local packages
	pattern := "./..."
	if len(cfg.args) > 0 {
		pattern = cfg.args[0]
	}

	packages, err := internal.ListPackages(".", pattern)
	if err != nil {
		return nil, err
	}

	defs := []*model.Definition{}
	getDef := func(in *model.Definition) *model.Definition {
		for _, def := range defs {
			if def.Package.Equal(in.Package) {
				return def
			}
		}
		return nil
	}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, cfg.verbose)
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

func lint(cfg *options) error {
	var lintErrors []error

	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	for _, def := range defs {
		_, importCollisions := def.Imports.Map()
		for _, err := range importCollisions {
			lintErrors = append(lintErrors, err)
		}
	}

	if len(lintErrors) == 0 {
		return nil
	}

	for _, err := range lintErrors {
		fmt.Println(err)
	}

	return errors.New("Linter not passing")
}
