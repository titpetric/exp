package lint

import (
	"os"

	"golang.org/x/exp/slices"
)

// Run is the entrypoint for `go-fsck lint`.
func Run() (err error) {
	cfg := NewOptions()

	if slices.Contains(os.Args, "help") {
		cfg.PrintHelp()
		return nil
	}

	return lint(cfg)
}
