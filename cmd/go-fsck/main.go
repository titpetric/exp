package main

import (
	"fmt"
	"os"

	"github.com/titpetric/exp/cmd/go-fsck/coverage"
	"github.com/titpetric/exp/cmd/go-fsck/diff"
	"github.com/titpetric/exp/cmd/go-fsck/docs"
	"github.com/titpetric/exp/cmd/go-fsck/extract"
	"github.com/titpetric/exp/cmd/go-fsck/internal/help"
	"github.com/titpetric/exp/cmd/go-fsck/query"
	"github.com/titpetric/exp/cmd/go-fsck/report"
	"github.com/titpetric/exp/cmd/go-fsck/restore"
	"github.com/titpetric/exp/cmd/go-fsck/search"
	"github.com/titpetric/exp/cmd/go-fsck/stats"
)

func main() {
	if err := start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func start() (err error) {
	commands := map[string]func() error{
		"extract":  extract.Run,
		"coverage": coverage.Run,
		"restore":  restore.Run,
		"stats":    stats.Run,
		"search":   search.Run,
		"query":    query.Run,
		"docs":     docs.Run,
		"diff":     diff.Run,
		"report":   report.Run,
	}

	// A command line with no command, or one asking for the help page, gets
	// the page listing the commands.
	if len(os.Args) < 2 {
		return help.Write(os.Stdout, helpSpec())
	}
	switch os.Args[1] {
	case "help", "--help", "-h":
		return help.Write(os.Stdout, helpSpec())
	}

	commandFn, ok := commands[os.Args[1]]
	if ok {
		return commandFn()
	}

	return fmt.Errorf("Unknown command: %q", os.Args[1])
}

// helpSpec is the page the command prints.
func helpSpec() help.Spec {
	return help.Spec{
		Name:    "go-fsck",
		Tagline: "code introspection over a data model of Go source",
		Usage: []string{
			"go-fsck <command> [flags]",
			"go-fsck <command> --help",
		},
		Description: `A Go tree is read into a document of packages and the declarations they
hold: types, consts, vars and funcs, each with its doc comment, its imports
and the symbols it reaches. The document is JSON, it is written by extract,
and every other command reads it.

The default document is go-fsck.json in the working directory. A command
reading one takes --input-file, and a command that finds no document parses
the tree in front of it instead.

The model, the parser behind it and the linters over it are splint, a module
of its own. What go-fsck adds is the commands below.`,
		Commands: []help.Command{
			{Name: "coverage", About: "fold a coverage profile into the model, per function and per package"},
			{Name: "diff", About: "compare the exported API and the go.mod of two documents"},
			{Name: "docs", About: "render the document as markdown, a spec, an import list or plantuml"},
			{Name: "extract", About: "read a Go tree into a go-fsck.json document"},
			{Name: "query", About: "find functions by the signature they are declared with"},
			{Name: "report", About: "report which test functions reach which symbols"},
			{Name: "restore", About: "write the document back out as source, one symbol per file"},
			{Name: "search", About: "find functions by name, or by the symbol they reference"},
			{Name: "stats", About: "godoc coverage, package size, import usage and reverse symbol usage"},
		},
		Examples: []help.Example{
			{Command: "go-fsck extract -r ./...", About: "read the tree into go-fsck.json"},
			{Command: "go-fsck docs > docs/api.md", About: "the API reference of what was read"},
			{Command: "go-fsck stats", About: "what the packages weigh and what they reach for"},
			{Command: "go-fsck diff --old old.json --new new.json", About: "what a release takes away"},
			{Command: "go-fsck restore -i model.json -o internal/model", About: "the document back out as source"},
			{Command: "go-fsck extract --help", About: "the page for one command"},
		},
		Notes: `Every command takes help, --help or -h and prints its own page.`,
	}
}
