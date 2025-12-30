package loader

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"

	"github.com/titpetric/exp/cmd/go-fsck/internal/collector"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// Load definitions from package located in sourcePath.
func Load(in *model.Package, includeTests bool, verbose bool) ([]*model.Definition, error) {
	if in.Pkg == nil {
		return nil, fmt.Errorf("No pkg present in package: %s", in)
	}

	pkg := in.Pkg
	sourcePath := in.Path

	//fset := token.NewFileSet()
	fset := in.Pkg.Fset

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: includeTests,
		Fset:  fset,
	}
	_ = cfg

	if verbose {
		log.Printf("Loading package %s %q", sourcePath, in.Name())
	}

	//pkgs, err := parser.ParseDir(fset, sourcePath, nil, parser.ParseComments)
	//	pkgs, err := packages.Load(cfg, sourcePath)
	//	if err != nil {
	//		return nil, err
	//	}

	files := []*ast.File{}
	//	for _, pkg := range pkgs {
	for _, file := range pkg.Syntax {
		filename := path.Base(fset.Position(file.Pos()).Filename)
		if !strings.HasSuffix(filename, ".go") {
			// skip test packages that don't end in .go
			continue
		}

		src, err := os.ReadFile(path.Join(sourcePath, filename))
		if err != nil {
			return nil, fmt.Errorf("Error reading in source file: %s", filename)
		}

		if tags := BuildTags(src); len(tags) > 0 {
			// skipped files with build tags
			continue
		}

		files = append(files, file)
		//		}
	}

	sink := collector.NewCollector(fset)

	insp := inspector.New(files)
	insp.WithStack(nil, sink.Visit)

	results := sink.Clean(verbose)

	// Attach the package information to all returned definitions
	for _, def := range results {
		def.Pkg = pkg
		def.ImportPath = in.ImportPath
		def.ID = in.ID
	}

	return results, nil
}

// ReadFile loads the definitions from a json file
func ReadFile(inputPath string) ([]*model.Definition, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	var result []*model.Definition
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	for _, decl := range result {
		decl.Fill()
	}

	return result, nil
}
