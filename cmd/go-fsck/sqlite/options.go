package sqlite

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

type options struct {
	inputFile string

	filter    string
	exclude   string
	reference string

	full    bool
	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")

	flag.StringVar(&cfg.filter, "filter", cfg.filter, "filter imports that match (sql LIKE)")
	flag.StringVar(&cfg.exclude, "exclude", cfg.exclude, "exclude imports that match (sql NOT LIKE)")

	flag.BoolVar(&cfg.full, "full", cfg.full, "resolve imports to full path")
	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s stats <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
