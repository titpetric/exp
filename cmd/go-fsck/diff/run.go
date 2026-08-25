package diff

import (
	"errors"
	"os"

	"golang.org/x/exp/slices"
)

// Run is the entrypoint for `go-fsck diff`.
func Run() (err error) {
	cfg := NewOptions()

	if slices.Contains(os.Args, "help") {
		PrintHelp()
		return nil
	}

	if cfg.oldFile == "" || cfg.newFile == "" {
		PrintHelp()
		return errors.New("both --old and --new are required")
	}

	return diff(cfg)
}
