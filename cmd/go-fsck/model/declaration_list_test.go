package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeclarationListSort_ReceiversFirst(t *testing.T) {
	// Test that functions without receivers come before functions with receivers
	// when sorted by receiver type.

	dl := DeclarationList{
		&Declaration{
			Kind:      FuncKind,
			Name:      "Zulu",
			Receiver:  "",
			Signature: "Zulu() string",
		},
		&Declaration{
			Kind:      FuncKind,
			Name:      "Alpha",
			Receiver:  "Ref",
			Signature: "Alpha() string",
		},
		&Declaration{
			Kind:      FuncKind,
			Name:      "Beta",
			Receiver:  "",
			Signature: "Beta() string",
		},
	}

	dl.Sort()

	// Expected order:
	// 1. Beta (no receiver, exported=false)
	// 2. Zulu (no receiver, exported=false)
	// 3. Alpha (Ref receiver)

	assert.Len(t, dl, 3)

	// Functions without receiver should come first
	assert.Empty(t, dl[0].Receiver, "First func should have no receiver")
	assert.Empty(t, dl[1].Receiver, "Second func should have no receiver")
	assert.NotEmpty(t, dl[2].Receiver, "Third func should have a receiver")
}

func TestDeclarationListSort_Stable(t *testing.T) {
	// Test that sorting is deterministic and produces same results each time

	makeDecl := func(name, receiver string) *Declaration {
		return &Declaration{
			Kind:      FuncKind,
			Name:      name,
			Receiver:  receiver,
			Signature: name + "() string",
		}
	}

	original := DeclarationList{
		makeDecl("String", "Ref"),
		makeDecl("Equal", "Declaration"),
		makeDecl("GetNames", "Declaration"),
		makeDecl("Find", ""),
		makeDecl("Walk", ""),
	}

	// Sort multiple times
	for i := 0; i < 5; i++ {
		dl := make(DeclarationList, len(original))
		copy(dl, original)
		dl.Sort()

		// Check order is consistent
		expectedOrder := []struct {
			name     string
			receiver string
		}{
			{"Find", ""},                // no receiver first
			{"Walk", ""},                // no receiver
			{"Equal", "Declaration"},    // receiver
			{"GetNames", "Declaration"}, // same receiver, sorted by name
			{"String", "Ref"},           // different receiver
		}

		for j, exp := range expectedOrder {
			assert.Equal(t, exp.name, dl[j].Name, "Iteration %d, position %d name mismatch", i, j)
			assert.Equal(t, exp.receiver, dl[j].Receiver, "Iteration %d, position %d receiver mismatch", i, j)
		}
	}
}
