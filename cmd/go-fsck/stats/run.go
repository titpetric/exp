package stats

import (
	"os"

	"golang.org/x/exp/slices"
)

// Run is the entrypoint for `go-fsck stats`.
func Run() (err error) {
	cfg := NewOptions()

	if slices.Contains(os.Args, "help") {
		PrintHelp()
		return nil
	}

	return stats(cfg)
}
