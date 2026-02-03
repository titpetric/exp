package edges

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_Memory(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	count, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestInsertSymbols(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	edges := []*Edge{
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyType",
			SymbolKind: TypeKind,
			IsExported: true,
			File:       "my_type.go",
			Line:       10,
		},
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyFunc",
			SymbolKind: FuncKind,
			IsExported: true,
			File:       "my_func.go",
			Line:       20,
		},
	}

	err = db.InsertSymbols(edges)
	assert.NoError(t, err)

	count, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestInsertSymbols_WithReceiver(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	edges := []*Edge{
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyType",
			SymbolKind: TypeKind,
			IsExported: true,
			File:       "my_type.go",
			Line:       10,
		},
		{
			ImportPath: "github.com/pkg",
			SymbolName: "Method",
			Receiver:   "MyType",
			SymbolKind: FuncKind,
			IsExported: true,
			File:       "my_type.go",
			Line:       20,
		},
	}

	err = db.InsertSymbols(edges)
	assert.NoError(t, err)

	count, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestInsertSymbols_Duplicate(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	edge := &Edge{
		ImportPath: "github.com/pkg",
		SymbolName: "MyType",
		SymbolKind: TypeKind,
		IsExported: true,
		File:       "my_type.go",
		Line:       10,
	}

	err = db.InsertSymbols([]*Edge{edge})
	assert.NoError(t, err)

	// Insert same edge again - should be ignored due to UNIQUE constraint
	err = db.InsertSymbols([]*Edge{edge})
	assert.NoError(t, err)

	count, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 1, count) // Still 1, not 2
}

func TestInsertRelationships(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// First insert symbols
	edges := []*Edge{
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyType",
			SymbolKind: TypeKind,
			IsExported: true,
			File:       "my_type.go",
			Line:       10,
		},
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyFunc",
			SymbolKind: FuncKind,
			IsExported: true,
			File:       "my_func.go",
			Line:       20,
		},
	}

	err = db.InsertSymbols(edges)
	require.NoError(t, err)

	// Insert relationship
	rel := &Relationship{
		From: edges[1], // MyFunc
		To:   edges[0], // MyType
		Type: ArgumentRel,
	}

	err = db.InsertRelationships([]*Relationship{rel})
	assert.NoError(t, err)

	count, err := db.RelationshipCount()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestInsertRelationships_MissingSymbol(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Insert only one symbol
	edge := &Edge{
		ImportPath: "github.com/pkg",
		SymbolName: "MyType",
		SymbolKind: TypeKind,
		IsExported: true,
		File:       "my_type.go",
		Line:       10,
	}

	err = db.InsertSymbols([]*Edge{edge})
	require.NoError(t, err)

	// Try to insert relationship to non-existent symbol
	missingEdge := &Edge{
		ImportPath: "github.com/pkg",
		SymbolName: "NonExistent",
		SymbolKind: FuncKind,
		IsExported: true,
		File:       "missing.go",
		Line:       99,
	}

	rel := &Relationship{
		From: missingEdge,
		To:   edge,
		Type: ArgumentRel,
	}

	err = db.InsertRelationships([]*Relationship{rel})
	// Should not error, but relationship should not be inserted
	assert.NoError(t, err)

	count, err := db.RelationshipCount()
	assert.NoError(t, err)
	assert.Equal(t, 0, count) // No relationship inserted
}

func TestInsertAll(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	edges := []*Edge{
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyType",
			SymbolKind: TypeKind,
			IsExported: true,
			File:       "my_type.go",
			Line:       10,
		},
		{
			ImportPath: "github.com/pkg",
			SymbolName: "MyFunc",
			SymbolKind: FuncKind,
			IsExported: true,
			File:       "my_func.go",
			Line:       20,
		},
	}

	rel := &Relationship{
		From: edges[1],
		To:   edges[0],
		Type: ArgumentRel,
	}

	err = db.InsertAll(edges, []*Relationship{rel})
	assert.NoError(t, err)

	symbolCount, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 2, symbolCount)

	relCount, err := db.RelationshipCount()
	assert.NoError(t, err)
	assert.Equal(t, 1, relCount)
}

func TestInsertAll_Empty(t *testing.T) {
	db, err := NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = db.InsertAll([]*Edge{}, []*Relationship{})
	assert.NoError(t, err)

	symbolCount, err := db.SymbolCount()
	assert.NoError(t, err)
	assert.Equal(t, 0, symbolCount)
}
