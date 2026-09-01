package restore

import (
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
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

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")
	flag.StringVarP(&cfg.outputPath, "output-path", "o", cfg.outputPath, "output `PATH`")
	flag.BoolVar(&cfg.split, "split", cfg.split, "one file per symbol, unexported ones included")
	flag.StringSliceVar(&cfg.keep, "keep", cfg.keep, "declaration `KINDS` to write: type, const, var, func")
	flag.BoolVar(&cfg.removeUnexported, "remove-unexported", cfg.removeUnexported, "remove unexported symbols")
	flag.BoolVar(&cfg.noTests, "no-tests", cfg.noTests, "do not restore test files")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")

	flag.Usage = PrintHelp
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

// PrintHelp writes the page for `go-fsck restore`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck restore",
		Tagline: "write the document back out as source, one symbol per file",
		Usage: []string{
			"go-fsck restore [flags]",
		},
		Description: `Every package of the document is written as source under the output path, at
the path it was extracted from, so a document of one package restores into one
directory and a document of a tree restores as the tree.

Each symbol goes in the file named for it: a type in the file of its name, its
methods and its constructor beside it, a const typed by that type with it, the
unexported half in the file named for the package, and the values belonging to
no type in const.go and vars.go. --split puts every symbol in a file of its
own, the unexported ones included.

Each file imports what its own declarations reach and nothing else, and
everything is written through goimports. Nothing else is rewritten: the
package clause and the import paths are what the source said, so a package
reaching another package of the same tree still reaches it by the path it was
published under.

What it is for is embedding a package inside your own module rather than
beside it in go.mod, which is something go.mod cannot express.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck restore -i model.json -o internal/model", About: "one package, into a directory of your module"},
			{Command: "go-fsck restore -i oida.json -o vendor/oida --no-tests", About: "a whole tree, without its test files"},
			{Command: "go-fsck restore -i model.json -o internal/model --keep type,const --remove-unexported", About: "the exported types and consts alone"},
		},
		Notes: `The document has to carry sources, which is what extract --include-sources
is for, and a document extracted without them has nothing to write.

Anything that is not a Go declaration is not in the document, so a package
embedding files with go:embed restores without them.`,
	}
}
