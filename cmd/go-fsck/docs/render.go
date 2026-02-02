package docs

import (
	"encoding/json"
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// Read the exported go-fsck.json data.
	defs, err := loader.ReadFile(cfg.inputFile)
	if err == nil {
		return defs, nil
	}

	packagePath := "./..."
	if len(cfg.args) > 1 {
		// [docs .]
		packagePath = cfg.args[1]
	}

	// list current local packages
	packages, err := internal.ListPackages(".", packagePath)
	if err != nil {
		return nil, err
	}

	defs = []*model.Definition{}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, cfg.verbose)
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

func render(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	if cfg.split {
		return renderSplit(cfg, defs)
	}

	switch cfg.render {
	case "spec":
		return renderSpec(cfg, defs)
	case "imports":
		return renderImports(cfg, defs)
	case "json":
		return renderJSON(cfg, defs)
	case "puml", "plantuml":
		return renderPlantUML(cfg, defs)
	default:
		return renderMarkdown(cfg, defs)
	}
}

func renderJSON(_ *options, defs []*model.Definition) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(defs)
}
