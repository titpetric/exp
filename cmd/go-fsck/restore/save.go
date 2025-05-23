package restore

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func saveLayout(cfg *options, files map[string]model.DeclarationList, filenames []string) error {
	for _, filename := range filenames {
		if cfg.removeTests && strings.HasSuffix(filename, "_test.go") {
			if cfg.verbose {
				fmt.Println("Skipping test:", filename)
			}
			continue
		}
		if cfg.keepTestsOnly && !strings.HasSuffix(filename, "_test.go") {
			if cfg.verbose {
				fmt.Println("Skipping non-test:", filename)
			}
			continue
		}

		decls := files[filename]
		decls.Sort()

		// Collect sources
		lines := []string{"package " + cfg.packageName}

		imports := map[string]bool{}
		for _, decl := range decls {
			for _, v := range decl.Imports {
				imports[v] = true
			}
		}
		if cfg.addDotImport != "" {
			imports[fmt.Sprintf(`%s "%s"`, ".", cfg.addDotImport)] = true
		}
		if len(imports) > 0 {
			lines = append(lines, "", "import (")
			lines = append(lines, maps.Keys(imports)...)
			lines = append(lines, "", ")")
		}

		for _, decl := range decls {
			if decl.Source != "" {
				lines = append(lines, "", decl.Source)
				continue
			}
			return fmt.Errorf("Missing source for object, %#v", decl)
		}

		if len(lines) > 0 {
			body := []byte(strings.Join(lines, "\n"))
			if err := os.WriteFile(filename, body, 0644); err != nil {
				return fmt.Errorf("error saving: %w", err)
			}
		}
	}

	return nil
}
