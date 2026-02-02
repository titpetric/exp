package docs

import (
	"fmt"
	"path"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func renderMarkdown(_ *options, defs []*model.Definition) error {
	// Loop through function definitions and collect referenced
	// symbols from imported packages. Globals may also reference
	// imported packages so this is incomplete at the moment.
	for _, def := range defs {
		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}

		var (
			types  = def.Types.Exported()
			consts = def.Consts.Exported()
			vars   = def.Vars.Exported()
			funcs  = def.Funcs.Exported()
		)

		var packageName = def.Package.Path // strings.ReplaceAll(def.Package.ImportPath, "github.com/", "")
		if packageName == "." {
			packageName = path.Base(def.Package.ImportPath)
		}

		fmt.Println("# Package", packageName)
		fmt.Println()
		fmt.Println("```go")
		fmt.Printf("import (\n\t\"%s\"\n}\n", def.Package.ImportPath)
		fmt.Println("```")

		if def.Doc != "" {
			fmt.Println(strings.TrimSpace(def.Doc))
			fmt.Println()
		}

		if len(types) > 0 {
			fmt.Println("## Types")
			fmt.Println()
			for _, v := range types {
				src := strings.TrimSpace(v.Source)
				fmt.Printf("```go\n%s\n```\n\n", src)
			}
		}

		if len(consts) > 0 {
			fmt.Println("## Consts")
			fmt.Println()
			for _, v := range consts {
				src := strings.TrimSpace(v.Source)
				fmt.Printf("```go\n%s\n```\n\n", src)
			}
		}
		if len(vars) > 0 {
			fmt.Println("## Vars")
			fmt.Println()
			for _, v := range vars {
				src := strings.TrimSpace(v.Source)
				fmt.Printf("```go\n%s\n```\n\n", src)
			}
		}

		symbol := func(fn *model.Declaration) string {
			if fn.Receiver != "" {
				return "func (" + fn.Receiver + ") " + fn.Signature
			}
			return "func " + fn.Signature
		}

		if len(funcs) > 0 {
			for {
				fmt.Println("## Function symbols")
				fmt.Println()

				for _, fn := range funcs {
					fmt.Printf("- `%s`\n", symbol(fn))
				}
				fmt.Println()

				// Documented functions first.
				for _, fn := range funcs {
					if fn.Doc == "" {
						continue
					}

					fmt.Printf("### %s\n\n", fn.Name)
					fmt.Println(strings.TrimSpace(fn.Doc))
					fmt.Println()
					fmt.Printf("```go\n%s\n```\n\n", symbol(fn))
				}

				// List undocumented ones.
				for _, fn := range funcs {
					if fn.Doc != "" {
						continue
					}

					fmt.Printf("### %s\n\n", fn.Name)
					fmt.Printf("```go\n%s\n```\n\n", symbol(fn))
				}
				fmt.Println()
				break
			}
		}
	}

	return nil
}
