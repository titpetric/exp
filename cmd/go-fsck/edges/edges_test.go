package edges

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

func TestExtractEdges_Types(t *testing.T) {
	def := &model.Definition{
		Package: model.Package{
			Package:    "pkg",
			ImportPath: "github.com/pkg",
		},
		Types: model.DeclarationList{
			&model.Declaration{
				Kind:   model.TypeKind,
				Name:   "MyType",
				File:   "my_type.go",
				Line:   10,
				Type:   "struct",
				Source: "type MyType struct {}",
			},
		},
	}

	edges, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	assert.Len(t, edges, 1)
	assert.Len(t, rels, 0)
	assert.Equal(t, "MyType", edges[0].SymbolName)
	assert.Equal(t, TypeKind, edges[0].SymbolKind)
}

func TestExtractEdges_Funcs(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Funcs: model.DeclarationList{
			&model.Declaration{
				Kind:   model.FuncKind,
				Name:   "MyFunc",
				File:   "my_func.go",
				Line:   20,
				Source: "func MyFunc() {}",
			},
		},
	}

	edges, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	assert.Len(t, edges, 1)
	assert.Len(t, rels, 0)
	assert.Equal(t, "MyFunc", edges[0].SymbolName)
	assert.Equal(t, FuncKind, edges[0].SymbolKind)
	assert.Empty(t, edges[0].Receiver)
}

func TestExtractEdges_MethodWithReceiver(t *testing.T) {
	typeEdge := &model.Declaration{
		Kind:   model.TypeKind,
		Name:   "MyType",
		File:   "my_type.go",
		Line:   10,
		Source: "type MyType struct {}",
	}

	methodEdge := &model.Declaration{
		Kind:     model.FuncKind,
		Name:     "Method",
		Receiver: "MyType",
		File:     "my_type.go",
		Line:     20,
		Source:   "func (t MyType) Method() {}",
	}

	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Types: model.DeclarationList{
			typeEdge,
		},
		Funcs: model.DeclarationList{
			methodEdge,
		},
	}

	edges, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	assert.Len(t, edges, 2)
	assert.Len(t, rels, 1)

	// Check relationship
	assert.Equal(t, "Method", rels[0].From.SymbolName)
	assert.Equal(t, "MyType", rels[0].To.SymbolName)
	assert.Equal(t, ReceiverRel, rels[0].Type)
}

func TestExtractEdges_ArgumentTypes(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Types: model.DeclarationList{
			&model.Declaration{
				Kind:   model.TypeKind,
				Name:   "MyType",
				File:   "my_type.go",
				Line:   10,
				Source: "type MyType struct {}",
			},
		},
		Funcs: model.DeclarationList{
			&model.Declaration{
				Kind:      model.FuncKind,
				Name:      "MyFunc",
				File:      "my_func.go",
				Line:      20,
				Arguments: []string{"MyType"},
				Source:    "func MyFunc(t MyType) {}",
			},
		},
	}

	e, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	require.NotNil(t, e)
	require.Len(t, e, 2)

	// Find the argument relationship
	argRels := findRelsByType(rels, ArgumentRel)
	assert.Len(t, argRels, 1)
	assert.Equal(t, "MyFunc", argRels[0].From.SymbolName)
	assert.Equal(t, "MyType", argRels[0].To.SymbolName)
}

func TestExtractEdges_ReturnTypes(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Types: model.DeclarationList{
			&model.Declaration{
				Kind:   model.TypeKind,
				Name:   "MyType",
				File:   "my_type.go",
				Line:   10,
				Source: "type MyType struct {}",
			},
		},
		Funcs: model.DeclarationList{
			&model.Declaration{
				Kind:    model.FuncKind,
				Name:    "MyFunc",
				File:    "my_func.go",
				Line:    20,
				Returns: []string{"MyType"},
				Source:  "func MyFunc() MyType { return MyType{} }",
			},
		},
	}

	e, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	require.NotNil(t, e)

	retRels := findRelsByType(rels, ReturnRel)
	assert.Len(t, retRels, 1)
	assert.Equal(t, "MyFunc", retRels[0].From.SymbolName)
	assert.Equal(t, "MyType", retRels[0].To.SymbolName)
}

func TestExtractEdges_TestFunctionInference(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Types: model.DeclarationList{
			&model.Declaration{
				Kind:   model.TypeKind,
				Name:   "MyType",
				File:   "my_type.go",
				Line:   10,
				Source: "type MyType struct {}",
			},
		},
		Funcs: model.DeclarationList{
			&model.Declaration{
				Kind:   model.FuncKind,
				Name:   "TestMyType",
				File:   "my_type_test.go",
				Line:   20,
				Source: "func TestMyType(t *testing.T) {}",
			},
		},
	}

	e, rels, err := ExtractEdges(def)

	require.NoError(t, err)
	require.NotNil(t, e)

	testRels := findRelsByType(rels, TestRel)
	assert.Len(t, testRels, 1)
	assert.Equal(t, "TestMyType", testRels[0].From.SymbolName)
	assert.Equal(t, "MyType", testRels[0].To.SymbolName)
}

func TestParseTypeReference(t *testing.T) {
	tests := []struct {
		typeRef  string
		expected string
	}{
		{"string", ""},          // builtin
		{"MyType", "MyType"},    // simple type
		{"*MyType", "MyType"},   // pointer
		{"[]int", ""},           // builtin slice
		{"[]*MyType", "MyType"}, // slice of pointers
		{"map[string]MyType", "MyType"},
		{"http.Handler", "Handler"},
		{"error", ""}, // builtin interface
	}

	for _, tt := range tests {
		t.Run(tt.typeRef, func(t *testing.T) {
			result := parseTypeReference(tt.typeRef)
			require.Equal(t, tt.expected, result, "parseTypeReference(%q) should return %q, got %q", tt.typeRef, tt.expected, result)
		})
	}
}

func TestInferTestTarget(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"TestMyType", "MyType"},
		{"TestMyFunc_Case1", "MyFunc"},
		{"Test_myFunc", ""}, // unexported target
		{"TestValue", "Value"},
		{"Test", ""},     // no target
		{"NotATest", ""}, // not a test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferTestTarget(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractEdges_Nil(t *testing.T) {
	edges, rels, err := ExtractEdges(nil)

	assert.Error(t, err)
	assert.Nil(t, edges)
	assert.Nil(t, rels)
}

func TestExtractEdges_Vars(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Vars: model.DeclarationList{
			&model.Declaration{
				Kind:   model.VarKind,
				Name:   "myVar",
				File:   "my_var.go",
				Line:   10,
				Source: "var myVar = 0",
			},
		},
	}

	edges, _, err := ExtractEdges(def)

	require.NoError(t, err)
	assert.Len(t, edges, 1)
	assert.Equal(t, VarKind, edges[0].SymbolKind)
}

func TestExtractEdges_Consts(t *testing.T) {
	def := &model.Definition{
		Package: newTestPackage("github.com/pkg"),
		Consts: model.DeclarationList{
			&model.Declaration{
				Kind:   model.ConstKind,
				Name:   "MyConst",
				File:   "const.go",
				Line:   10,
				Source: "const MyConst = 42",
			},
		},
	}

	edges, _, err := ExtractEdges(def)

	require.NoError(t, err)
	assert.Len(t, edges, 1)
	assert.Equal(t, ConstKind, edges[0].SymbolKind)
}

// Helper functions

func newTestPackage(importPath string) model.Package {
	parts := strings.Split(importPath, "/")
	pkgName := parts[len(parts)-1]
	return model.Package{
		Package:    pkgName,
		ImportPath: importPath,
	}
}

func findRelsByType(rels []*Relationship, relType RelationshipType) []*Relationship {
	var result []*Relationship
	for _, rel := range rels {
		if rel.Type == relType {
			result = append(result, rel)
		}
	}
	return result
}
