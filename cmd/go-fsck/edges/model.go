package edges

import (
	"fmt"
	"strings"
)

// SymbolKind represents the type of Go symbol.
type SymbolKind string

const (
	TypeKind  SymbolKind = "type"
	VarKind   SymbolKind = "var"
	ConstKind SymbolKind = "const"
	FuncKind  SymbolKind = "func"
)

// RelationshipType represents the nature of a relationship between symbols.
type RelationshipType string

const (
	ReceiverRel RelationshipType = "receiver" // Function has receiver type
	ArgumentRel RelationshipType = "argument" // Function parameter uses type
	ReturnRel   RelationshipType = "return"   // Function returns type
	UsesRel     RelationshipType = "uses"     // Function body references symbol
	TestRel     RelationshipType = "test"     // Test function covers symbol
)

// Edge represents a single symbol definition in the codebase.
//
// Symbols are uniquely identified by (ImportPath, SymbolName, Receiver).
// For methods, Receiver is non-empty (e.g., "MyType" or "*MyType").
// For non-methods, Receiver is empty.
type Edge struct {
	// ImportPath is the full import path (e.g., "github.com/owner/pkg").
	ImportPath string

	// SymbolName is the name of the symbol (e.g., "MyType", "myFunc").
	SymbolName string

	// Receiver is non-empty for methods, containing the receiver type name.
	// Example: "MyType" or "*MyType"
	Receiver string

	// SymbolKind is the kind of symbol (type, var, const, func).
	SymbolKind SymbolKind

	// IsExported is true if the symbol is exported (capitalized).
	IsExported bool

	// File is the source file path (relative, with .go extension).
	File string

	// Line is the line number where the symbol is defined.
	Line int
}

// FullName returns the fully qualified name of the symbol.
// For methods, it includes the receiver (e.g., "MyType.Method").
// For non-methods, it returns just the symbol name.
func (e *Edge) FullName() string {
	if e.Receiver != "" {
		return fmt.Sprintf("%s.%s", e.Receiver, e.SymbolName)
	}
	return e.SymbolName
}

// SymbolID returns the unique symbol identifier.
// Format: "import_path#SymbolName" or "import_path#Receiver.SymbolName" for methods.
func (e *Edge) SymbolID() string {
	if e.Receiver != "" {
		return fmt.Sprintf("%s#%s.%s", e.ImportPath, e.Receiver, e.SymbolName)
	}
	return fmt.Sprintf("%s#%s", e.ImportPath, e.SymbolName)
}

// Relationship represents a connection between two symbols.
type Relationship struct {
	// From is the symbol that initiates the relationship (e.g., a function).
	From *Edge

	// To is the target symbol (e.g., a type that is used).
	To *Edge

	// Type is the nature of the relationship (receiver, argument, return, uses, test).
	Type RelationshipType

	// Details is optional metadata as JSON string (e.g., {"index": 0} for argument position).
	Details string
}

// String returns a human-readable representation of the relationship.
func (r *Relationship) String() string {
	return fmt.Sprintf("%s -[%s]-> %s", r.From.SymbolID(), r.Type, r.To.SymbolID())
}

// ParseSymbolID parses a symbol ID into import path and symbol name.
// Returns (importPath, symbolName, receiver, error).
// Format: "import_path#SymbolName" or "import_path#Receiver.SymbolName"
func ParseSymbolID(symbolID string) (importPath string, symbolName string, receiver string, err error) {
	parts := strings.Split(symbolID, "#")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid symbol ID format: %s", symbolID)
	}

	importPath = parts[0]
	rest := parts[1]

	// Check if it's a method (contains dot)
	if strings.Contains(rest, ".") {
		methodParts := strings.SplitN(rest, ".", 2)
		receiver = methodParts[0]
		symbolName = methodParts[1]
	} else {
		symbolName = rest
	}

	return importPath, symbolName, receiver, nil
}
