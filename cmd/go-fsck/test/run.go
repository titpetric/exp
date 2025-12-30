package test

import (
	"os"
	"slices"
)

// Run is the entrypoint for `go-fsck test`.
func Run() error {
	args := os.Args[2:] // Skip "go-fsck" and "test"

	if slices.Contains(args, "help") {
		PrintHelp()
		return nil
	}

	// Check if -c flag is present
	hasCompile := slices.Contains(args, "-c")

	if !hasCompile {
		return runSingleTest(args)
	}

	return compileMultiModuleTests(args)
}
