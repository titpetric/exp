package lint

import (
	"fmt"
	"os"
	"path"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
)

type options struct {
	verbose   bool
	summarize bool
	jsonOut   bool
	rules     []string
	exclude   []string
	args      []string

	fs *internal.FlagSet
}

// NewOptions parses command-line flags and returns the lint options.
func NewOptions() *options {
	cfg := &options{
		rules: []string{"imports", "godoc", "func-args"},
	}

	cfg.fs = internal.NewFlagSet("lint")
	cfg.fs.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	cfg.fs.BoolVarP(&cfg.summarize, "summarize", "", cfg.summarize, "summarize linter issues instead of raw logs")
	cfg.fs.BoolVarP(&cfg.jsonOut, "json", "", cfg.jsonOut, "output results as JSON")
	cfg.fs.StringSliceVarP(&cfg.rules, "rules", "", cfg.rules, "linter rules to run")
	cfg.fs.StringSliceVarP(&cfg.exclude, "exclude", "", cfg.exclude, "linter rules to exclude")

	cfg.args = internal.ParseArgs(cfg.fs)

	return cfg
}

// GetRules returns the active rules after applying exclusions.
func (o *options) GetRules() []string {
	if len(o.exclude) == 0 {
		return o.rules
	}

	result := make([]string, 0, len(o.rules))
	for _, rule := range o.rules {
		var excluded bool
		for _, ex := range o.exclude {
			if ex == rule {
				excluded = true
				break
			}
		}
		if !excluded {
			result = append(result, rule)
		}
	}
	return result
}

// PrintHelp displays usage information for the lint command.
func (o *options) PrintHelp() {
	fmt.Printf("Usage: %s lint <options>:\n\n", path.Base(os.Args[0]))
	o.fs.PrintDefaults()
}
