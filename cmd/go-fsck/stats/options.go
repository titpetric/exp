package stats

import (
	"os"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
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

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")

	flag.StringVar(&cfg.filter, "filter", cfg.filter, "filter imports that match `PATTERN` (sql LIKE)")
	flag.StringVar(&cfg.exclude, "exclude", cfg.exclude, "exclude imports that match `PATTERN` (sql NOT LIKE)")

	flag.BoolVar(&cfg.full, "full", cfg.full, "resolve imports to full path")
	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	return cfg
}

// PrintHelp writes the page for `go-fsck stats`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck stats",
		Tagline: "what the packages weigh and what they reach for",
		Usage: []string{
			"go-fsck stats [flags]",
		},
		Description: `Four reports are written to stdout, one after another, each as a markdown
section.

Documentation is how much of each package carries a doc comment. Package
stats is what each package declares, counted by kind. Import usage is which
import each package pulls in and how often. Reverse symbol usage is the other
direction: which packages reach a given import, and which symbols of it they
name, which is what says how much a dependency is actually used.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck stats", About: "the four reports for the document beside the tree"},
			{Command: "go-fsck stats -i model.json", About: "the same for another document"},
		},
	}
}
