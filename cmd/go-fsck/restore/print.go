package restore

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func printLayout(cfg *options, files map[string]model.DeclarationList, filenames []string) error {
	for _, filename := range filenames {
		if cfg.removeTests && strings.HasSuffix(filename, "_test.go") {
			continue
		}
		if cfg.keepTestsOnly && !strings.HasSuffix(filename, "_test.go") {
			continue
		}
		decls := files[filename]

		lines := []string{}
		for _, decl := range decls {
			if cfg.removeUnexported {
				if decl.Name != "" && !ast.IsExported(decl.Name) {
					continue
				}
			}

			if decl.Kind == model.FuncKind {
				receiver := strings.TrimLeft(decl.Receiver, "*")
				if receiver != "" {
					if cfg.removeUnexported && !ast.IsExported(receiver) {
						continue
					}

					lines = append(lines, fmt.Sprintf("- func (%s) %s", decl.Receiver, decl.Signature))
					continue
				}
				lines = append(lines, "- func "+decl.Signature)
				continue
			}
			if len(decl.Names) > 0 {
				for _, name := range decl.Names {
					if cfg.removeUnexported {
						if name != "" && !ast.IsExported(name) {
							continue
						}
					}
					lines = append(lines, "- "+decl.Kind.String()+" "+name)
				}
				continue
			}
			lines = append(lines, "- "+decl.Kind.String()+" "+decl.Name)
		}

		if cfg.statsFiles {
			_ = json.NewEncoder(os.Stdout).Encode(struct {
				File    string
				Count   int
				Symbols []string
			}{filename, len(lines), lines})
			continue
		}

		if len(lines) > 0 {
			fmt.Println("#", len(lines), filename)
			fmt.Println()
			fmt.Println(strings.Join(lines, "\n"))
			fmt.Println()
		}
	}

	return nil
}
