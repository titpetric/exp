package coverage

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

type options struct {
	inputFile    string
	outputFile   string
	coverageFile string

	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	flag.StringVarP(&cfg.outputFile, "output-file", "o", cfg.outputFile, "output file")
	flag.StringVarP(&cfg.coverageFile, "coverage-file", "c", cfg.coverageFile, "summary coverage file")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s search <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
