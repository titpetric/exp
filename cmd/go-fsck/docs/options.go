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
	model  bool
	hide   string

	docs bool

	verbose bool
	args    []string
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
		render:    "markdown",
		docs:      false,
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	flag.StringVar(&cfg.render, "render", cfg.render, "print results as [markdown, json, ...]")
	flag.StringVar(&cfg.focus, "focus", cfg.focus, "focus on configured symbol")
	flag.BoolVar(&cfg.model, "model", cfg.model, "model mode: skip functions and interfaces")
	flag.StringVar(&cfg.hide, "hide", cfg.hide, "comma-separated list of types to hide")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	cfg.args = flag.Args()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s docs <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
