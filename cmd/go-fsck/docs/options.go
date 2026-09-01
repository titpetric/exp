package docs

import (
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
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

// defaults is the options a run starts from.
func defaults() *options {
	return &options{
		inputFile: "go-fsck.json",
		render:    "markdown",
		docs:      false,
		out:       ".",
	}
}

// flags is the command line, registered on a parser of its own. The help page
// walks one of these, so a flag is documented by having been defined.
func (cfg *options) flags() *internal.FlagSet {
	fs := internal.NewFlagSet("docs")
	fs.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")
	fs.StringVar(&cfg.render, "render", cfg.render, "print results as `FORMAT` [markdown, json, ...]")
	fs.StringVar(&cfg.focus, "focus", cfg.focus, "focus on configured `SYMBOL`")
	fs.BoolVar(&cfg.model, "model", cfg.model, "model mode: skip functions and interfaces")
	fs.StringVar(&cfg.hide, "hide", cfg.hide, "comma-separated `LIST` of types to hide")
	fs.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	fs.BoolVar(&cfg.split, "split", cfg.split, "split output file per package")
	fs.StringVar(&cfg.out, "out", cfg.out, "output `DIR` (used with --split)")
	fs.StringVar(&cfg.strip, "strip", cfg.strip, "`PREFIX` to strip from import path for filename")
	return fs
}

// NewOptions parses command-line flags and returns the docs options.
func NewOptions() *options {
	cfg := defaults()

	cfg.fs = cfg.flags()
	cfg.fs.Usage = PrintHelp

	cfg.args = internal.ParseArgs(cfg.fs)

	return cfg
}

// PrintHelp writes the page for `go-fsck docs`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck docs",
		Tagline: "render the document as an API reference",
		Usage: []string{
			"go-fsck docs [flags]",
		},
		Description: `The document is rendered to stdout: the package godoc, the types, consts
and vars it declares, and the signature of every exported function. Function
bodies are not printed, whether or not the document carries sources.

Godoc examples are the exception, and are printed whole under an Examples
heading, each wrapped in a section named after the function. They live in a
test package, so the document has to have been extracted with both
--include-sources and --include-tests to hold them.

--render picks what is written: markdown is the default, spec is the
declarations alone, imports and puml are plantuml diagrams of what the
packages reach, and json is the document itself.

--split writes one file per package under --out instead of one document on
stdout, named after the import path with --strip taken off the front.`,
		Flags: defaults().flags(),
		Examples: []help.Example{
			{Command: "go-fsck docs > docs/api.md", About: "the API reference of the document beside the tree"},
			{Command: "go-fsck docs --render puml", About: "a plantuml diagram of the types"},
			{Command: "go-fsck docs --split --out docs/api --strip github.com/titpetric/exp", About: "one markdown file per package"},
		},
	}
}
