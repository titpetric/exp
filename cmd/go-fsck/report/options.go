package report

import (
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	inputFile string

	json    bool
	verbose bool
	args    []string

	fs *internal.FlagSet
}

// defaults is the options a run starts from.
func defaults() *options {
	return &options{
		inputFile: "go-fsck.json",
	}
}

// flags is the command line, registered on a parser of its own. The help page
// walks one of these, so a flag is documented by having been defined.
func (cfg *options) flags() *internal.FlagSet {
	fs := internal.NewFlagSet("report")
	fs.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")
	fs.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	fs.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	return fs
}

func NewOptions() *options {
	cfg := defaults()

	cfg.fs = cfg.flags()
	cfg.fs.Usage = PrintHelp

	cfg.args = internal.ParseArgs(cfg.fs)

	return cfg
}

// PrintHelp writes the page for `go-fsck report`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck report",
		Tagline: "report which test functions reach which symbols",
		Usage: []string{
			"go-fsck report [flags]",
		},
		Description: `Every test function of the document is walked and the symbols it references
are resolved to the package they come from. A test function is one named
Test, Benchmark or Example, one taking a testing argument, or any function
declared in a _test.go file.

One line is written per reference: the test function, the package and symbol
it reaches, and whether that package is external to the one the test sits in.
What it is for is reading which symbol a test touches, which is what says
whether a symbol is tested at all and by what.

The document has to have been extracted with --include-tests, or there is
nothing here to walk.`,
		Flags: defaults().flags(),
		Examples: []help.Example{
			{Command: "go-fsck report", About: "the references of the document beside the tree"},
			{Command: "go-fsck report -i model.json", About: "the references of another document"},
		},
	}
}
