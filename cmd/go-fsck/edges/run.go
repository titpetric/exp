package edges

import (
	"fmt"
	"os"

	"golang.org/x/exp/slices"
)

// Run is the entrypoint for `go-fsck edges`.
func Run() error {
	cfg := NewOptions()

	if slices.Contains(os.Args, "help") {
		cfg.PrintHelp()
		return nil
	}

	return runEdges(cfg)
}

func runEdges(cfg *Options) error {
	return fmt.Errorf("edges command not yet implemented")
}
