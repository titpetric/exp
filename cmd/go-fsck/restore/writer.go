package restore

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal/files"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	. "github.com/titpetric/exp/cmd/go-fsck/model"
)

// RestoreDefinition creates files for the given definition in the package.
func RestoreDefinition(definition *Definition, cfg *options) error {
	var (
		basePath         = cfg.outputPath
		removeUnexported = cfg.removeUnexported
	)

	// Ensure the base directory exists
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	var modelDeclarations []*Declaration
	var richModelDeclarations []*Declaration
	var modelTestDeclarations []*Declaration
	var richModelTestDeclarations []*Declaration

	modelImports := make(StringSet)
	richModelImports := make(StringSet)
	modelTestImports := make(StringSet)
	richModelTestImports := make(StringSet)

	// Collect declarations and imports
	for _, decl := range definition.Types {
		if removeUnexported && !ast.IsExported(decl.Name) {
			continue
		}

		imports := definition.Imports[decl.File]

		if strings.HasSuffix(decl.File, "_test.go") {
			if decl.SelfContained {
				modelTestDeclarations = append(modelTestDeclarations, decl)
				mergeImports(&modelTestImports, imports)
			} else {
				richModelTestDeclarations = append(richModelTestDeclarations, decl)
				mergeImports(&richModelTestImports, imports)
			}
		} else {
			if decl.SelfContained {
				modelDeclarations = append(modelDeclarations, decl)
				mergeImports(&modelImports, imports)
			} else {
				richModelDeclarations = append(richModelDeclarations, decl)
				mergeImports(&richModelImports, imports)
			}
		}
	}

	// Write self-contained declarations to model.go
	if err := writeFile(definition.Package, modelDeclarations, modelImports, filepath.Join(basePath, "model.go")); err != nil {
		return err
	}

	// Write non-self-contained declarations to model_rich.go
	if err := writeFile(definition.Package, richModelDeclarations, richModelImports, filepath.Join(basePath, "model_rich.go")); err != nil {
		return err
	}

	// Write self-contained test declarations to model_test.go
	if err := writeFile(definition.Package, modelTestDeclarations, modelTestImports, filepath.Join(basePath, "model_test.go")); err != nil {
		return err
	}

	// Write non-self-contained test declarations to model_rich_test.go
	if err := writeFile(definition.Package, richModelTestDeclarations, richModelTestImports, filepath.Join(basePath, "model_rich_test.go")); err != nil {
		return err
	}

	return nil
}

// mergeImports merges a list of imports into the given import set
func mergeImports(importsSet *StringSet, imports []string) {
	for _, imp := range imports {
		importsSet.Add(imp, imp)
	}
}

// writeFile writes the declarations and imports to a specified file using the files package
func writeFile(pkg model.Package, declarations []*Declaration, imports StringSet, filePath string) error {
	// Collect types
	var types []string
	for _, decl := range declarations {
		types = append(types, decl.Source)
	}

	// Create a File struct
	file := &files.File{
		Filename: filePath,
		Package:  pkg.Name(),
		Imports:  imports.All(),
		Types:    types,
	}

	// Write the file content to disk
	return file.Flush()
}
