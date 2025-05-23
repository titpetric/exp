package extract

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	flag "github.com/spf13/pflag"
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
	flag.StringVarP(&cfg.outputFile, "output-file", "o", cfg.outputFile, "output file")
	flag.StringVarP(&cfg.sourcePath, "source-path", "i", cfg.sourcePath, "source path")
	flag.BoolVar(&cfg.includeTests, "include-tests", cfg.includeTests, "include test files")
	flag.BoolVar(&cfg.includeSources, "include-sources", cfg.includeSources, "include sources")
	flag.BoolVar(&cfg.prettyJSON, "pretty-json", cfg.prettyJSON, "print pretty json")
	flag.BoolVarP(&cfg.recursive, "recursive", "r", cfg.recursive, "recurse packages")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	cfg.outputFile, _ = filepath.Abs(cfg.outputFile)

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s extract <sourcePath> <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
