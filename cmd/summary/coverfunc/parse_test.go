package coverfunc

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/summary/internal"
)

//go:embed testdata
var coverfunc_testdata embed.FS

func TestParse(t *testing.T) {
	f, err := coverfunc_testdata.Open("testdata/cover.txt")
	assert.NoError(t, err)

	lines, err := internal.ReadFields(f)
	assert.NoError(t, err)

	result := Parse(lines)

	t.Logf("%#v", result[0])
}

// TestSkipUncoveredDilutes holds the --skip-uncovered contract: uncovered
// functions are dropped from the output, not from the aggregation, so a file
// keeps the average its uncovered functions dilute.
func TestSkipUncoveredDilutes(t *testing.T) {
	lines := [][]string{
		{"pkg/a.go:3:", "covered", "100.0%"},
		{"pkg/a.go:9:", "uncovered", "0.0%"},
		{"pkg/b.go:3:", "dead", "0.0%"},
	}

	parsed := Parse(lines)
	assert.Len(t, parsed, 3)

	files := ByFile(parsed)
	assert.Len(t, files, 2)
	for _, f := range files {
		if f.Filename == "pkg/a.go" {
			assert.Equal(t, 50.0, f.Coverage, "uncovered function must dilute the file average")
		}
	}
}
