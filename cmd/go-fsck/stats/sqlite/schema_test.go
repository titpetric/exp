package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatements(t *testing.T) {
	data := Statements()
	assert.True(t, len(data) > 0, "expected at least one schema migration")
}
