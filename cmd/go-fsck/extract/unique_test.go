package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// Helper function to create a Definition instance
func newDefinition(importPath string) *model.Definition {
	return &model.Definition{
		Package: model.Package{
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
		expected []*model.Definition
	}{
		{
			name: "No duplicates",
			input: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
			},
			expected: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
			},
		},
		{
			name: "Duplicates with merge",
			input: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg1"), // Duplicate
				newDefinition("pkg2"),
			},
			expected: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
			},
		},
		{
			name: "Multiple duplicates",
			input: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
				newDefinition("pkg1"), // Duplicate
				newDefinition("pkg3"),
				newDefinition("pkg2"), // Duplicate
			},
			expected: []*model.Definition{
				newDefinition("pkg1"),
				newDefinition("pkg2"),
				newDefinition("pkg3"),
			},
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unique(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
