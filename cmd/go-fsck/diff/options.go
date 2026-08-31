package diff

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
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

	flag.StringVar(&cfg.oldFile, "old", cfg.oldFile, "go-fsck.json of the older revision")
	flag.StringVar(&cfg.newFile, "new", cfg.newFile, "go-fsck.json of the newer revision")

	flag.BoolVar(&cfg.includeInternal, "include-internal", cfg.includeInternal, "compare internal packages as well")
	flag.BoolVar(&cfg.includeIndirect, "include-indirect", cfg.includeIndirect, "compare indirect go.mod requirements as well")
	flag.BoolVar(&cfg.includeUnexported, "include-unexported", cfg.includeUnexported, "compare unexported declarations and internal packages as well, reported but never breaking")

	flag.BoolVar(&cfg.json, "json", cfg.json, "print results as json")
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.Parse()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s diff --old <file> --new <file> <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
