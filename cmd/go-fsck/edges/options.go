package edges

import (
	"flag"
	"fmt"
	"os"
)

// Options holds configuration for the edges command.
type Options struct {
	InputFile  string
	OutputFile string
	Verbose    bool
}

// NewOptions parses command-line flags and returns Options.
func NewOptions() *Options {
	opts := &Options{}

	fs := flag.NewFlagSet("edges", flag.ContinueOnError)
	fs.StringVar(&opts.InputFile, "i", "model/restored/go-fsck.json", "Input JSON file from extraction")
	fs.StringVar(&opts.OutputFile, "o", "edges.db", "Output database file (or :memory: for in-memory)")
	fs.BoolVar(&opts.Verbose, "v", false, "Verbose output")

	_ = fs.Parse(os.Args[2:])

	return opts
}

// PrintHelp prints the help message.
func (opts *Options) PrintHelp() {
	fmt.Print(`Usage: go-fsck edges [options]

Extract symbol edges and relationships from a go-fsck model.

Options:
  -i string
      Input JSON file from go-fsck extract (default "model/restored/go-fsck.json")
  -o string
      Output database file or ":memory:" for in-memory database (default "edges.db")
  -v
      Verbose output
  -h, -help
      Show this help message
`)
}
