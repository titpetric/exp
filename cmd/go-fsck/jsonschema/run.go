package jsonschema

import (
	"os"
	"slices"
)

// Run is the entrypoint for `go-fsck jsonschema`.
func Run() (err error) {
	cfg := NewOptions()

	if slices.Contains(os.Args, "help") {
		PrintHelp()
		return nil
	}

	return ParseAndConvertStruct(cfg)
}
