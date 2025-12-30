package lint

import (
	"fmt"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

type options struct {
	verbose   bool
	summarize bool
	jsonOut   bool
	rules     []string
	exclude   []string
	args      []string
}

func NewOptions() *options {
	cfg := &options{
		rules: []string{"godoc"},
	}
	flag.BoolVarP(&cfg.verbose, "verbose", "v", cfg.verbose, "verbose output")
	flag.BoolVarP(&cfg.summarize, "summarize", "", cfg.summarize, "summarize linter issues instead of raw logs")
	flag.BoolVarP(&cfg.jsonOut, "json", "", cfg.jsonOut, "output results as JSON")
	flag.StringSliceVarP(&cfg.rules, "rules", "", cfg.rules, "linter rules to run")
	flag.StringSliceVarP(&cfg.exclude, "exclude", "", cfg.exclude, "linter rules to exclude")
	flag.Parse()

	cfg.args = flag.Args()

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

func PrintHelp() {
	fmt.Printf("Usage: %s lint <options>:\n\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}
