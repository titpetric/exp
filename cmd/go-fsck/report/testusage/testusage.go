package testusage

import (
	"fmt"
	"log"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

type FuncRef struct {
	FuncDecl          *model.Declaration
	Definition        *model.Definition // Reference to the Definition to access imports and other data
	ReferencedPackage string
	ReferencedSymbol  string
	ImportPath        string
	IsExternal        bool
}

func (f *FuncRef) String() string {
	return fmt.Sprintf("%s --> %s.%s (external %v)", f.FuncDecl.Name, f.ReferencedPackage, f.ReferencedSymbol, f.IsExternal)
}

type Report struct {
	Refs []*FuncRef
}

func (r *Report) String() string {
	var b strings.Builder
	for _, v := range r.Refs {
		b.WriteString(v.String() + "\n")
	}
	return b.String()
}

// NewReport analyzes the provided definitions, resolves external dependencies, and returns a list of *FuncRef.
func NewReport(definitions []*model.Definition) (*Report, error) {
	var allFuncRefs []*FuncRef

	for _, def := range definitions {
		log.Printf("Processing package: %s\n", def.Package.Package)
		funcRefs, err := getFunctionNames(def)
		if err != nil {
			return nil, err
		}
		allFuncRefs = append(allFuncRefs, funcRefs...)
	}

	// Resolve external dependencies
	resolvedRefs, err := resolveExternalDependencies(allFuncRefs, definitions)
	if err != nil {
		return nil, err
	}

	log.Printf("Total test function references found: %d\n", len(resolvedRefs))
	return &Report{resolvedRefs}, nil
}

// isTestFunction checks if a function is a test function based on its name or signature.
func isTestFunction(funcDecl *model.Declaration) bool {
	switch {
	case strings.HasPrefix(funcDecl.Name, "Test"):
		return true
	case strings.HasPrefix(funcDecl.Name, "Benchmark"):
		return true
	case strings.HasPrefix(funcDecl.Name, "Example"):
		return true
	}

	// Simplified argument type check
	for _, arg := range funcDecl.Arguments {
		if strings.HasPrefix(arg, "testing.") {
			return true
		}
	}

	return false
}

// getFunctionNames processes a definition and returns a slice of *FuncRef.
func getFunctionNames(def *model.Definition) ([]*FuncRef, error) {
	var funcRefs []*FuncRef

	importMap, _ := def.Imports.Map()

	for _, funcDecl := range def.Funcs {
		if isTestFunction(funcDecl) || strings.HasSuffix(funcDecl.File, "_test.go") {
			for prodPkg, symbols := range funcDecl.References {
				for _, symbol := range symbols {
					fullImport, ok := importMap[prodPkg]
					if !ok {
						return nil, fmt.Errorf("no import reference for: %s", prodPkg)
					}

					// this check excludes stdlib couplings
					// but we may want them to discover
					// things like mutex use, etc.
					if !strings.Contains(fullImport, ".") {
						continue
					}

					isExternal := !strings.HasPrefix(fullImport, def.Package.ImportPath)
					ref := &FuncRef{
						FuncDecl:          funcDecl,
						Definition:        def,
						ReferencedPackage: prodPkg,
						ReferencedSymbol:  symbol,
						ImportPath:        fullImport,
						IsExternal:        isExternal,
					}
					funcRefs = append(funcRefs, ref)
				}
			}
		}
	}

	return funcRefs, nil
}

// resolveExternalDependencies processes funcRefs and resolves the correct import paths for external symbols.
func resolveExternalDependencies(funcRefs []*FuncRef, definitions []*model.Definition) ([]*FuncRef, error) {
	for _, ref := range funcRefs {
		symbolPackage := ref.ReferencedPackage

		// First, try to resolve within the same definition
		imported, _ := ref.Definition.Imports.Map()
		if len(imported) == 0 {
			continue
		}

		matchedImport, ok := imported[symbolPackage]
		if !ok {
			return nil, fmt.Errorf("no matching import found for symbol package: %s in function: %s", symbolPackage, ref.FuncDecl.Name)
		}

		ref.ImportPath = matchedImport
	}

	return funcRefs, nil
}
