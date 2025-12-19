package extract

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// Helper function to create a Definition instance
var defID int

func newDefinition(importPath string) *model.Definition {
	defID++
	return &model.Definition{
		Package: model.Package{
			ID:         fmt.Sprintf("pkg%d", defID),
			ImportPath: importPath,
		},
	}
}

// Test the unique function
func TestUnique(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    []*model.Definition
		expected int
	}{
		{
			name: "No duplicates",
			input: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
			},
			expected: 2,
		},
		{
			name: "Duplicates with merge",
			input: []*model.Definition{
				{Package: model.Package{ID: "same", ImportPath: "pkg1"}},
				{Package: model.Package{ID: "same", ImportPath: "pkg1"}}, // Duplicate
				newDefinition("pkg2"),
			},
			expected: 2,
		},
		{
			name: "Multiple duplicates",
			input: []*model.Definition{
				{Package: model.Package{ID: "id1", ImportPath: "pkg1"}},
				{Package: model.Package{ID: "id2", ImportPath: "pkg2"}},
				{Package: model.Package{ID: "id1", ImportPath: "pkg1"}}, // Duplicate
				newDefinition("pkg3"),
				{Package: model.Package{ID: "id2", ImportPath: "pkg2"}}, // Duplicate
			},
			expected: 3,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unique(tt.input)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}
