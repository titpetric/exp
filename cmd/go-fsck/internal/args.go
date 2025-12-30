package internal

import (
	"os"

	flag "github.com/spf13/pflag"
)

// FlagSet is an alias for pflag.FlagSet to avoid requiring direct pflag imports.
type FlagSet = flag.FlagSet

// NewFlagSet creates a new FlagSet for a subcommand.
// It automatically handles the argument offset, skipping the program name
// and command name when parsing (os.Args[2:]).
func NewFlagSet(command string) *FlagSet {
	return flag.NewFlagSet(command, flag.ExitOnError)
}

// ParseArgs parses the command-line arguments for a subcommand.
// It skips os.Args[0] (program) and os.Args[1] (command name).
func ParseArgs(fs *FlagSet) []string {
	args := os.Args[2:]
	if len(os.Args) < 3 {
		args = []string{}
	}
	fs.Parse(args)
	return fs.Args()
}
