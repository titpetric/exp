package docs

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
)

type options struct {
	inputFile string

	render string
	focus  string
	model  bool
	hide   string

	docs bool

	verbose bool
	split   bool
	out     string
	strip   string
	args    []string

	fs *internal.FlagSet
}

// NewOptions parses command-line flags and returns the docs options.
func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
		render:    "markdown",
		docs:      false,
		out:       ".",
	}

	cfg.fs = internal.NewFlagSet("docs")
	cfg.fs.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	cfg.fs.StringVar(&cfg.render, "render", cfg.render, "print results as [markdown, json, ...]")
	cfg.fs.StringVar(&cfg.focus, "focus", cfg.focus, "focus on configured symbol")
	cfg.fs.BoolVar(&cfg.model, "model", cfg.model, "model mode: skip functions and interfaces")
	cfg.fs.StringVar(&cfg.hide, "hide", cfg.hide, "comma-separated list of types to hide")
	cfg.fs.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	cfg.fs.BoolVar(&cfg.split, "split", cfg.split, "split output file per package")
	cfg.fs.StringVar(&cfg.out, "out", cfg.out, "output directory (used with --split)")
	cfg.fs.StringVar(&cfg.strip, "strip", cfg.strip, "prefix to strip from import path for filename")

	cfg.args = internal.ParseArgs(cfg.fs)

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s docs <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
