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

	result := Parse(lines, false)

	t.Logf("%#v", result[0])
}
