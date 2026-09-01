package query

import (
	"os"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	inputFile string

	showHandlers   bool
	showMiddleware bool

	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	flag.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input `FILE`")

	flag.BoolVar(&cfg.showHandlers, "handlers", cfg.showHandlers, "show http handlers")
	flag.BoolVar(&cfg.showMiddleware, "middleware", cfg.showMiddleware, "show tyk middleware")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	return cfg
}

// PrintHelp writes the page for `go-fsck query`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck query",
		Tagline: "find functions by the signature they are declared with",
		Usage: []string{
			"go-fsck query [flags]",
		},
		Description: `Functions grouped by the signature they share are functions that could be
one package, and this is what finds them. Two signatures are known by name.

--handlers keeps the functions taking http.ResponseWriter and *http.Request,
which is a net/http handler. --middleware keeps the functions taking those
two and a third argument and returning an error and an int, which is the tyk
middleware signature. Given neither, every function of the document is
listed.

The result is a markdown table by default, and the declarations themselves
with --json.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck query --handlers", About: "the http handlers of the document"},
			{Command: "go-fsck query --middleware", About: "the middleware of the document"},
			{Command: "go-fsck query --handlers --json", About: "the same declarations, for a program to read"},
		},
	}
}
