package search

import (
	"os"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	inputFile string

	name      string
	reference string

	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")

	flag.StringVar(&cfg.name, "name", cfg.name, "function `NAME` match (case sensitive)")
	flag.StringVar(&cfg.reference, "reference", cfg.reference, "reference `SYMBOL` (e.g. 'oas', or 'oas.OAS')")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	return cfg
}

// PrintHelp writes the page for `go-fsck search`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck search",
		Tagline: "find functions by name, or by the symbol they reference",
		Usage: []string{
			"go-fsck search [flags]",
		},
		Description: `--name keeps the functions whose name contains the string, and the match is
case sensitive. --reference keeps the functions that reach a package: the
value is a package name, or a package name and a symbol, and only the package
half of it is matched.

Given both, a function has to satisfy both. Given neither, every function of
the document is listed. The result is a markdown table of the function, the
file it is in and the symbols it reached, and the declarations themselves
with --json.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck search --name Handler", About: "every function with Handler in its name"},
			{Command: "go-fsck search --reference oas.OAS", About: "every function reaching the oas package"},
			{Command: "go-fsck search --name New --reference sqlx --json", About: "both, for a program to read"},
		},
	}
}
