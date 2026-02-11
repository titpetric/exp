package jsonschema

import (
	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	workingDir, packagePath := ".", "."

	packages, err := internal.ListPackages(workingDir, packagePath)
	if err != nil {
		return nil, err
	}

	defs := []*model.Definition{}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, false)
		if err != nil {
			return nil, err
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

	return defs, nil
}
