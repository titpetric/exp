package docs

import (
	"fmt"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func renderImports(opt *options, defs []*model.Definition) error {
	pkgs := make(map[string]model.Package)
	for _, def := range defs {
		pkgs[def.Package.ImportPath] = def.Package
		fmt.Println(def.Package.ImportPath)
	}

	fmt.Println("@startuml")

	imports := model.NewStringSet()

	fmt.Println("' Have", len(defs), "definitions")

	for _, def := range defs {
		decls := def.DeclarationList()

		fmt.Println("' Have", len(decls), "declarations in", def.Package.ImportPath)
		fmt.Println()

		importMap, _ := def.Imports.Map(def.Imports.All())

		for _, long := range importMap {
			if !strings.Contains(long, ".") {
				continue
			}

			pkg, ok := pkgs[long]
			if !ok {
				continue
			}

			addImport(&imports, def.Package, pkg)
		}
	}

	for _, src := range imports.Keys() {
		uses := imports[src]
		for _, use := range uses {
			fmt.Printf("[%s] --|> [%s]\n", src, use)
		}
	}

	fmt.Println("@enduml")

	return nil
}

func addImport(s *model.StringSet, src, dest model.Package) {
	from, to := sanitize(src.ImportPath), sanitize(dest.ImportPath)
	if from == "" {
		from = src.Package
	}
	if to == "" {
		to = src.Package
	}
	s.Add(from, to)
}

func sanitize(n string) string {
	n = strings.Trim(n, "./")
	return n
}
