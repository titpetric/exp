package docs

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

type options struct {
	inputFile string

	render string
	focus  string

	verbose bool
	args    []string
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
		render:    "markdown",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	flag.StringVar(&cfg.render, "render", cfg.render, "print results as [markdown, json, ...]")
	flag.StringVar(&cfg.focus, "focus", cfg.focus, "focus on configured symbol")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	cfg.args = flag.Args()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s docs <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
