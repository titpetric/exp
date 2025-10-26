package coverfunc

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

type options struct {
	GroupByFiles   bool
	GroupByPackage bool
	SkipUncovered  bool

	RenderJSON bool
}

func NewOptions() *options {
	cfg := &options{}

	flag.BoolVarP(&cfg.GroupByFiles, "files", "f", cfg.GroupByFiles, "Group coverage by file")
	flag.BoolVarP(&cfg.GroupByPackage, "packages", "p", cfg.GroupByPackage, "Group coverage by package")
	flag.BoolVar(&cfg.SkipUncovered, "skip-uncovered", cfg.SkipUncovered, "Skip uncovered files")
	flag.BoolVar(&cfg.RenderJSON, "json", false, "Render output as json")
	flag.Parse()

	return cfg
}

func PrintHelp() {
	fmt.Printf("Usage: %s coverfunc <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
