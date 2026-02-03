package edges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdge_FullName(t *testing.T) {
	tests := []struct {
		name     string
		edge     *Edge
		expected string
	}{
		{
			name: "non-method symbol",
			edge: &Edge{
				SymbolName: "MyType",
				Receiver:   "",
			},
			expected: "MyType",
		},
		{
			name: "method with value receiver",
			edge: &Edge{
				SymbolName: "Method",
				Receiver:   "MyType",
			},
			expected: "MyType.Method",
		},
		{
			name: "method with pointer receiver",
			edge: &Edge{
				SymbolName: "Method",
				Receiver:   "*MyType",
			},
			expected: "*MyType.Method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.edge.FullName())
		})
	}
}

func TestEdge_SymbolID(t *testing.T) {
	tests := []struct {
		name     string
		edge     *Edge
		expected string
	}{
		{
			name: "type symbol",
			edge: &Edge{
				ImportPath: "github.com/pkg",
				SymbolName: "MyType",
				Receiver:   "",
			},
			expected: "github.com/pkg#MyType",
		},
		{
			name: "function symbol",
			edge: &Edge{
				ImportPath: "github.com/pkg",
				SymbolName: "MyFunc",
				Receiver:   "",
			},
			expected: "github.com/pkg#MyFunc",
		},
		{
			name: "method symbol",
			edge: &Edge{
				ImportPath: "github.com/pkg",
				SymbolName: "MyMethod",
				Receiver:   "MyType",
			},
			expected: "github.com/pkg#MyType.MyMethod",
		},
		{
			name: "pointer receiver method",
			edge: &Edge{
				ImportPath: "github.com/pkg",
				SymbolName: "String",
				Receiver:   "*Ref",
			},
			expected: "github.com/pkg#*Ref.String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.edge.SymbolID())
		})
	}
}

func TestParseSymbolID(t *testing.T) {
	tests := []struct {
		name         string
		symbolID     string
		expectedPath string
		expectedName string
		expectedRecv string
		shouldErr    bool
	}{
		{
			name:         "simple type",
			symbolID:     "github.com/pkg#MyType",
			expectedPath: "github.com/pkg",
			expectedName: "MyType",
			expectedRecv: "",
		},
		{
			name:         "method",
			symbolID:     "github.com/pkg#MyType.Method",
			expectedPath: "github.com/pkg",
			expectedName: "Method",
			expectedRecv: "MyType",
		},
		{
			name:         "pointer receiver",
			symbolID:     "github.com/pkg#*Ref.String",
			expectedPath: "github.com/pkg",
			expectedName: "String",
			expectedRecv: "*Ref",
		},
		{
			name:      "invalid format",
			symbolID:  "invalid",
			shouldErr: true,
		},
		{
			name:      "too many parts",
			symbolID:  "a#b#c",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, name, recv, err := ParseSymbolID(tt.symbolID)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
				assert.Equal(t, tt.expectedName, name)
				assert.Equal(t, tt.expectedRecv, recv)
			}
		})
	}
}

func TestRelationship_String(t *testing.T) {
	rel := &Relationship{
		From: &Edge{
			ImportPath: "github.com/pkg",
			SymbolName: "MyFunc",
		},
		To: &Edge{
			ImportPath: "github.com/pkg",
			SymbolName: "MyType",
		},
		Type: ArgumentRel,
	}

	expected := "github.com/pkg#MyFunc -[argument]-> github.com/pkg#MyType"
	assert.Equal(t, expected, rel.String())
}
