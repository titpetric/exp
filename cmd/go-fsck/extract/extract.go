package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/internal/telemetry"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func loadModuleTree(ctx context.Context, cfg *options, modules []internal.Module, pattern string) ([]*model.Definition, error) {
	result := []*model.Definition{}

	// Get absolute path of source for comparison
	absSourcePath, _ := filepath.Abs(cfg.sourcePath)

	for _, m := range modules {
		defs, err := walkPackage(ctx, m.Dir, pattern, cfg.includeTests, cfg.verbose)
		if err != nil {
			return nil, err
		}

		// The go.mod was already located by ListModules, so there is nothing
		// left to search for; a module whose go.mod will not parse is reported
		// without one, the same as a tree that has none.
		if module, err := internal.ReadModule(m.Filename); err == nil {
			setModule(defs, module)
		}

		// Adjust paths to be relative to root, not module directory
		moduleRelPath := strings.TrimPrefix(m.Dir, absSourcePath)
		if moduleRelPath != "" {
			moduleRelPath = strings.TrimPrefix(moduleRelPath, string(filepath.Separator))
			for _, def := range defs {
				pkgRelPath := strings.TrimPrefix(def.Package.Path, ".")
				def.Package.Path = "." + filepath.Join(moduleRelPath, pkgRelPath)
			}
		}

		result = append(result, defs...)
	}
	return result, nil
}

// setModule records the go.mod on every package extracted from it. The model
// is a flat list of packages, so a package that is asked what it builds
// against has to answer on its own.
func setModule(defs []*model.Definition, module *model.Module) {
	if module == nil {
		return
	}
	for _, def := range defs {
		def.Module = module
	}
}

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var pattern string
	if pattern = "."; cfg.recursive {
		pattern = "./..."
	}

	defs := []*model.Definition{}

	if pattern == "./..." {
		modules, err := internal.ListModules(cfg.sourcePath, pattern)
		if err != nil {
			return nil, err
		}

		d, err := loadModuleTree(ctx, cfg, modules, pattern)
		if err != nil {
			return nil, err
		}
		defs = append(defs, d...)
	}

	if pattern == "." {
		// Resolved before walking, since listing packages leaves the process
		// in the source directory and a relative source path stops resolving
		// to the same place afterwards.
		module, err := internal.FindModule(cfg.sourcePath)
		if err != nil {
			return nil, err
		}

		d, err := walkPackage(ctx, cfg.sourcePath, pattern, cfg.includeTests, cfg.verbose)
		if err != nil {
			return nil, err
		}
		setModule(d, module)
		defs = append(defs, d...)
	}

	defs = unique(defs)

	for _, def := range defs {
		if !cfg.includeSources {
			def.ClearSource()
		}
		if !def.TestPackage || !cfg.includeTests {
			def.ClearTestFiles()
		}
		if def.TestPackage {
			def.ClearNonTestFiles()
		}
	}

	return defs, nil
}

func walkPackage(ctx context.Context, sourcePath string, pattern string, includeTests bool, verbose bool) ([]*model.Definition, error) {
	defer telemetry.Start("extract.walkPackage " + sourcePath).End()
	defer runtime.GC()

	// fmt.Println("walking:", sourcePath, pattern, "tests", includeTests, "verbose", verbose)
	packages, err := internal.ListPackages(sourcePath, pattern)
	if err != nil {
		return nil, err
	}

	defs := []*model.Definition{}
	for _, pkg := range packages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if !includeTests {
			if pkg.TestPackage {
				continue
			}
		}

		//span := telemetry.Start("extract.walkPackage " + pkg.ImportPath)

		d, err := loader.Load(pkg, includeTests, false)
		if err != nil {
			return nil, err
		}

		// White box test include whole package scope. Lie.
		if pkg.TestPackage {
			if !strings.HasSuffix(pkg.Package, "_test") {
				pkg.Package += "_test"
				pkg.ImportPath += "_test" // More about the binary, it's test scope even if not black box.
			}
		}

		for _, v := range d {
			v.Package.ID = pkg.ID
			v.Package.ImportPath = pkg.ImportPath
			v.Package.Path = pkg.Path
			v.Package.Package = pkg.Package
			v.Package.TestPackage = pkg.TestPackage
		}

		defs = append(defs, d...)

		runtime.GC() // add some gc pressure

		//span.End()
	}
	return defs, nil
}

func extract(cfg *options) error {
	definitions, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	output := os.Stdout
	switch cfg.outputFile {
	case "", "-":
	default:
		fmt.Println(cfg.outputFile)
		var err error
		output, err = os.Create(cfg.outputFile)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(output)
	if cfg.prettyJSON {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(definitions)
}
