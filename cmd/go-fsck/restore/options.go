package restore

import (
	"fmt"
	"os"
	"path"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/tools/splint/model"
)

// options is what one restore was asked for.
type options struct {
	// inputFile is the document to read and outputPath the directory the
	// packages are written under, each at the path it was extracted from.
	inputFile  string
	outputPath string

	// split puts every symbol in the file named for it, unexported ones
	// included, rather than collecting the unexported half in one file.
	split bool

	// keep are the declaration kinds to write, and is every kind when empty.
	keep []string

	// removeUnexported drops what a consumer of the package cannot reach, and
	// noTests drops what the toolchain compiles into the test binary.
	removeUnexported bool
	noTests          bool

	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile:  "go-fsck.json",
		outputPath: ".",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	flag.StringVarP(&cfg.outputPath, "output-path", "o", cfg.outputPath, "output path")
	flag.BoolVar(&cfg.split, "split", cfg.split, "one file per symbol, unexported ones included")
	flag.StringSliceVar(&cfg.keep, "keep", cfg.keep, "declaration kinds to write: type, const, var, func")
	flag.BoolVar(&cfg.removeUnexported, "remove-unexported", cfg.removeUnexported, "remove unexported symbols")
	flag.BoolVar(&cfg.noTests, "no-tests", cfg.noTests, "do not restore test files")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")

	flag.Parse()

	return cfg
}

// kinds returns the declaration kinds to write, and whether the run was
// restricted to any.
func (o *options) kinds() (map[model.DeclarationKind]bool, error) {
	if len(o.keep) == 0 {
		return nil, nil
	}

	known := map[string]model.DeclarationKind{
		"type":  model.TypeKind,
		"const": model.ConstKind,
		"var":   model.VarKind,
		"func":  model.FuncKind,
	}

	out := map[model.DeclarationKind]bool{}
	for _, name := range o.keep {
		kind, ok := known[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("no such kind: %q (have type, const, var, func)", name)
		}
		out[kind] = true
	}

	return out, nil
}

func PrintHelp() {
	fmt.Printf("Usage: %s restore <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
	fmt.Print(`
Writes the packages of a document back out as source, one directory per
package under the output path, with each symbol in the file named for it.

  go-fsck extract -i github.com/titpetric/oida/model@main --include-sources -o model.json .
  go-fsck restore -i model.json -o internal/model

The document has to carry sources, which is what --include-sources is for.
`)
}
