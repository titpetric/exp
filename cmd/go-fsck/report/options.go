package report

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
)

type options struct {
	inputFile string

	json    bool
	verbose bool
	args    []string

	fs *internal.FlagSet
}

func NewOptions() *options {
	cfg := &options{
		inputFile: "go-fsck.json",
	}

	cfg.fs = internal.NewFlagSet("report")

	cfg.fs.StringVarP(&cfg.inputFile, "input-file", "i", cfg.inputFile, "input file")
	cfg.fs.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	cfg.fs.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")

	cfg.args = internal.ParseArgs(cfg.fs)

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s report <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
