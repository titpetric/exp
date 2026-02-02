package docs

import (
	"fmt"
	"path"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func renderSpec(opts *options, defs []*model.Definition) error {
	// Loop through function definitions and collect referenced
	// symbols from imported packages. Globals may also reference
	// imported packages so this is incomplete at the moment.
	for _, def := range defs {
		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "test") {
			continue
		}

		var (
			types  = def.Types.Exported()
			consts = def.Consts.Exported()
			vars   = def.Vars.Exported()
			funcs  = def.Funcs.Exported()
		)

		// all symbols if `-v` is given
		if opts.verbose {
			types = def.Types
			consts = def.Consts
			vars = def.Vars
			funcs = def.Funcs
		}

		var packageName = def.Package.Path // strings.ReplaceAll(def.Package.ImportPath, "github.com/", "")
		if packageName == "." {
			packageName = path.Base(def.Package.ImportPath)
		}

		fmt.Println("package:", packageName)
		fmt.Println("import:", def.Package.ImportPath)
		fmt.Println("symbols:")
		if len(types) > 0 {
			for _, v := range types {
				fmt.Println("- type:", v.GetNames())
			}
		}

		if len(consts) > 0 {
			for _, v := range consts {
				fmt.Println("- const:", v.GetNames())
			}
		}
		if len(vars) > 0 {
			for _, v := range vars {
				fmt.Println("- var:", v.GetNames())
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
				for _, fn := range funcs {
					fmt.Printf("- func: `%s`\n", symbol(fn))
				}
				break
			}
		}
	}

	return nil
}
