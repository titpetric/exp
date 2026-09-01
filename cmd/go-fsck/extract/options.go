package extract

import (
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	sourcePath string
	outputFile string

	includeTests   bool
	includeSources bool

	prettyJSON bool
	recursive  bool
	verbose    bool
}

func NewOptions() *options {
	cfg := &options{
		sourcePath: ".",
		outputFile: "go-fsck.json",
	}
	// handle: `go-fsck extract ./...`
	if len(os.Args) > 2 {
		if os.Args[len(os.Args)-1] == "./..." {
			cfg.recursive = true
		}
	}
	flag.StringVarP(&cfg.outputFile, "output-file", "o", cfg.outputFile, "output `FILE`")
	flag.StringVarP(&cfg.sourcePath, "source-path", "i", cfg.sourcePath, "source `PATH`")
	flag.BoolVar(&cfg.includeTests, "include-tests", cfg.includeTests, "include test files")
	flag.BoolVar(&cfg.includeSources, "include-sources", cfg.includeSources, "include sources")
	flag.BoolVar(&cfg.prettyJSON, "pretty-json", cfg.prettyJSON, "print pretty json")
	flag.BoolVarP(&cfg.recursive, "recursive", "r", cfg.recursive, "recurse packages")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	cfg.outputFile, _ = filepath.Abs(cfg.outputFile)

	return cfg
}

// PrintHelp writes the page for `go-fsck extract`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck extract",
		Tagline: "read a Go tree into a go-fsck.json document",
		Usage: []string{
			"go-fsck extract [flags] [pattern]",
		},
		Description: `Every package under the source path is parsed into a document of the
declarations it holds: types, consts, vars and funcs, each with its doc
comment, its position and the symbols it reaches. The document is what every
other command reads.

The trailing pattern is "./..." to read the whole tree, which is the same
thing --recursive says. One package is the default.

The source path is a directory of the tree you are in, or an import path with
a version after it. An import path is read out of the module cache, and
downloaded into it first if it is not there yet.

Sources and test files are left out unless they are asked for. A document
without sources cannot be restored, and one without tests holds no godoc
examples.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck extract -r ./...", About: "read the whole tree into go-fsck.json"},
			{Command: "go-fsck extract -i ./migrate --include-sources --include-tests", About: "one package, with its source and its tests"},
			{Command: "go-fsck extract -i github.com/titpetric/oida/model@main --include-sources -o model.json", About: "a package of another module, out of the module cache"},
		},
	}
}
