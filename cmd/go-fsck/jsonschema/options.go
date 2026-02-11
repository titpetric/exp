package jsonschema

import (
	"fmt"

	"github.com/spf13/pflag"
)

type options struct {
	sourcePath  string
	outputFile  string
	stripPrefix []string
}

func NewOptions() *options {
	cfg := &options{
		sourcePath:  ".",
		outputFile:  "schema.json",
		stripPrefix: []string{},
	}

	pflag.StringVarP(&cfg.sourcePath, "dir", "i", cfg.sourcePath, "Path to the directory that contains the root type (required)")
	pflag.StringVarP(&cfg.outputFile, "out", "o", cfg.outputFile, "Output file name (optional)")
	pflag.StringSliceVarP(&cfg.stripPrefix, "strip-prefix", "s", cfg.stripPrefix, "List of package prefixes to strip from definition names (optional)")
	pflag.Parse()

	return cfg
}

// PrintHelp prints usage information for your CLI.
func PrintHelp() {
	fmt.Println("Usage: go-fsck jsonschema [options]")
	fmt.Println()
	pflag.PrintDefaults()
}
