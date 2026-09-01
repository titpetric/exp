package diff

import (
	"os"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
)

type options struct {
	oldFile string
	newFile string

	includeInternal   bool
	includeIndirect   bool
	includeUnexported bool

	json    bool
	verbose bool
}

func NewOptions() *options {
	cfg := &options{}

	flag.StringVar(&cfg.oldFile, "old", cfg.oldFile, "`go-fsck.json` of the older revision")
	flag.StringVar(&cfg.newFile, "new", cfg.newFile, "`go-fsck.json` of the newer revision")

	flag.BoolVar(&cfg.includeInternal, "include-internal", cfg.includeInternal, "compare internal packages as well")
	flag.BoolVar(&cfg.includeIndirect, "include-indirect", cfg.includeIndirect, "compare indirect go.mod requirements as well")
	flag.BoolVar(&cfg.includeUnexported, "include-unexported", cfg.includeUnexported, "compare unexported declarations and internal packages as well, reported but never breaking")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Usage = PrintHelp
	flag.Parse()

	return cfg
}

// PrintHelp writes the page for `go-fsck diff`.
func PrintHelp() {
	help.Write(os.Stdout, helpSpec())
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck diff",
		Tagline: "compare the exported API and the go.mod of two documents",
		Usage: []string{
			"go-fsck diff --old <file> --new <file> [flags]",
		},
		Description: `Two documents are read and what happened between them is reported: to the
exported API, to the data model behind it, and to the go.mod the module is
built from. It answers the question a release has to answer before it is
tagged: does this take anything away?

A symbol is keyed by import path, receiver type and name, so moving a
declaration to another file, or adding a sibling to its const block, is not a
difference. A func carries its signature with the parameter names removed, so
renaming a parameter is not a change and changing its type is. A type carries
the shape it is declared with and its exported fields.

Removing or changing an exported symbol is breaking, and adding one is not.
Internal packages, indirect requirements and unexported declarations are left
out unless they are asked for, and what --include-unexported reports is never
breaking.

--verbose prints both signatures under a changed symbol, and --json writes
the whole comparison for a program to read.`,
		Flags: flag.CommandLine,
		Examples: []help.Example{
			{Command: "go-fsck diff --old old.json --new new.json", About: "what the newer revision takes away"},
			{Command: "go-fsck diff --old old.json --new new.json --verbose", About: "the same, with both sides of every change"},
			{Command: "go-fsck diff --old old.json --new new.json --json", About: "the comparison as data"},
		},
		Notes: `Both --old and --new are required, and each names a document extracted
with go-fsck extract.`,
	}
}
