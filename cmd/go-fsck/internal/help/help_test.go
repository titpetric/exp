package help

import (
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// TestWriteMarkdown renders a page to a writer that is not a terminal, which
// is the markdown rendering, and reads back every section and every flag.
func TestWriteMarkdown(t *testing.T) {
	fs := flag.NewFlagSet("fixture", flag.ContinueOnError)
	fs.StringP("input-file", "i", "go-fsck.json", "input `FILE`")
	fs.String("render", "markdown", "print results as `FORMAT`")
	fs.StringSlice("keep", nil, "declaration `KINDS` to write")
	fs.BoolP("verbose", "v", false, "verbose output")

	spec := Spec{
		Name:    "go-fsck fixture",
		Tagline: "a page written from a fixture",
		Usage:   []string{"go-fsck fixture [flags]"},
		Description: `The description is the paragraphs under the usage, and is written
out as it was given.`,
		Commands: []Command{
			{Name: "extract", About: "read a Go tree into a model"},
		},
		Flags: fs,
		Examples: []Example{
			{Command: "go-fsck fixture -i go-fsck.json", About: "read the document beside the tree"},
		},
		Notes: "The notes are what is left to say.",
	}

	out := &strings.Builder{}
	assert.NoError(t, Write(out, spec))
	page := out.String()

	sections := []string{
		"# go-fsck fixture",
		"a page written from a fixture",
		"## Usage",
		"go-fsck fixture [flags]",
		"The description is the paragraphs under the usage",
		"## Commands",
		"| `extract` | read a Go tree into a model |",
		"## Flags",
		"## Examples",
		"go-fsck fixture -i go-fsck.json",
		"The notes are what is left to say.",
	}
	for _, section := range sections {
		assert.Contains(t, page, section)
	}

	// Every flag is spelled the way it is typed: the letter in front of the
	// name where it has one, the name of its value after it where it takes
	// one, and the default where it is not the zero of its type.
	flags := []string{
		"| `-i, --input-file FILE` | `go-fsck.json` | input FILE |",
		"| `--render FORMAT` | `markdown` | print results as FORMAT |",
		"| `--keep KINDS` |  | declaration KINDS to write |",
		"| `-v, --verbose` |  | verbose output |",
	}
	for _, one := range flags {
		assert.Contains(t, page, one)
	}
}
