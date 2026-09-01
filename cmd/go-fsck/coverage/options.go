package coverage

import (
	"os"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	inputFile    string
	outputFile   string
	coverageFile string

	template string

	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")
	flag.StringVarP(&cfg.outputFile, "output-file", "o", cfg.outputFile, "output `FILE`")
	flag.StringVarP(&cfg.coverageFile, "coverage-file", "c", cfg.coverageFile, "summary coverage `FILE`")

	flag.StringVar(&cfg.template, "template", cfg.template, "Template `FILE` for the report")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	return cfg
}

// PrintHelp writes the page for `go-fsck coverage`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck coverage",
		Tagline: "fold a coverage profile into the model and report it",
		Usage: []string{
			"go-fsck coverage [flags]",
		},
		Description: `The coverage file is the summary a coverage run wrote: one entry per
function and one per package. Each entry is matched to the declaration it
names, by package, file and line, and the coverage it carries is written onto
the model beside the complexity that was measured there.

Without --output-file the report is printed: the function table as markdown,
or the same numbers as JSON with --json. With one, the model carrying the
coverage is written to that file instead, and the counts of what matched are
printed.

--verbose keeps the functions with no coverage in the report, which are left
out otherwise. --template names a text/template rendered with both tables,
.Functions and .Packages, in place of printing the function table alone.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck coverage -c coverage.json", About: "the coverage report, as a markdown table"},
			{Command: "go-fsck coverage -c coverage.json --json", About: "the same numbers, for a program to read"},
			{Command: "go-fsck coverage -c coverage.json -o covered.json", About: "write the model with the coverage folded in"},
		},
	}
}
