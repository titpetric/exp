package edges

import (
	"fmt"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ExtractEdges extracts symbol edges and relationships from a model Definition.
//
// Returns:
// - edges: all symbol definitions from the model
// - relationships: all relationships between symbols
// - error: if extraction fails
func ExtractEdges(def *model.Definition) ([]*Edge, []*Relationship, error) {
	if def == nil {
		return nil, nil, fmt.Errorf("definition is nil")
	}

	importPath := def.Package.Name()
	edges := make([]*Edge, 0)
	relationships := make([]*Relationship, 0)

	// Extract types
	for _, typeDecl := range def.Types {
		edge := &Edge{
			ImportPath: importPath,
			SymbolName: typeDecl.Name,
			SymbolKind: TypeKind,
			IsExported: typeDecl.IsExported(),
			File:       typeDecl.File,
			Line:       typeDecl.Line,
		}
		edges = append(edges, edge)
	}

	// Extract vars
	for _, varDecl := range def.Vars {
		edge := &Edge{
			ImportPath: importPath,
			SymbolName: varDecl.Name,
			SymbolKind: VarKind,
			IsExported: varDecl.IsExported(),
			File:       varDecl.File,
			Line:       varDecl.Line,
		}
		edges = append(edges, edge)
	}

	// Extract consts
	for _, constDecl := range def.Consts {
		names := constDecl.GetNames()
		for _, name := range names {
			edge := &Edge{
				ImportPath: importPath,
				SymbolName: name,
				SymbolKind: ConstKind,
				IsExported: constDecl.IsExported(),
				File:       constDecl.File,
				Line:       constDecl.Line,
			}
			edges = append(edges, edge)
		}
	}

	// Extract funcs
	for _, funcDecl := range def.Funcs {
		edge := &Edge{
			ImportPath: importPath,
			SymbolName: funcDecl.Name,
			Receiver:   funcDecl.Receiver,
			SymbolKind: FuncKind,
			IsExported: funcDecl.IsExported(),
			File:       funcDecl.File,
			Line:       funcDecl.Line,
		}
		edges = append(edges, edge)

		// Extract relationships for this function

		// 1. Receiver relationship
		if funcDecl.Receiver != "" {
			receiverType := strings.TrimLeft(funcDecl.Receiver, "*")
			receiverEdge := findEdgeByName(edges, receiverType)
			if receiverEdge != nil {
				relationships = append(relationships, &Relationship{
					From: edge,
					To:   receiverEdge,
					Type: ReceiverRel,
				})
			}
		}

		// 2. Argument type relationships
		for _, arg := range funcDecl.Arguments {
			argType := parseTypeReference(arg)
			if argType != "" {
				targetEdge := findEdgeInPackageOrExternal(edges, importPath, argType)
				if targetEdge != nil {
					relationships = append(relationships, &Relationship{
						From: edge,
						To:   targetEdge,
						Type: ArgumentRel,
					})
				}
			}
		}

		// 3. Return type relationships
		for _, ret := range funcDecl.Returns {
			retType := parseTypeReference(ret)
			if retType != "" {
				targetEdge := findEdgeInPackageOrExternal(edges, importPath, retType)
				if targetEdge != nil {
					relationships = append(relationships, &Relationship{
						From: edge,
						To:   targetEdge,
						Type: ReturnRel,
					})
				}
			}
		}

		// 4. Uses relationships (from References)
		for _, ref := range funcDecl.References.All() {
			refType := parseTypeReference(ref)
			if refType != "" && refType != funcDecl.Name {
				targetEdge := findEdgeByName(edges, refType)
				if targetEdge != nil {
					relationships = append(relationships, &Relationship{
						From: edge,
						To:   targetEdge,
						Type: UsesRel,
					})
				}
			}
		}

		// 5. Test relationships (infer from name)
		testTarget := inferTestTarget(funcDecl.Name)
		if testTarget != "" {
			targetEdge := findEdgeByName(edges, testTarget)
			if targetEdge != nil {
				relationships = append(relationships, &Relationship{
					From: edge,
					To:   targetEdge,
					Type: TestRel,
				})
			}
		}
	}

	return edges, relationships, nil
}

// parseTypeReference extracts the type name from a type reference string.
// Examples:
//
//	"string" -> "string"
//	"*MyType" -> "MyType"
//	"[]int" -> "int"
//	"[]*MyType" -> "MyType"
//	"map[string]MyType" -> "MyType"
func parseTypeReference(typeRef string) string {
	typeRef = strings.TrimSpace(typeRef)

	// Handle map types: map[K]V - take the V part
	if strings.HasPrefix(typeRef, "map[") {
		idx := strings.LastIndex(typeRef, "]")
		if idx != -1 && idx < len(typeRef)-1 {
			typeRef = typeRef[idx+1:]
			typeRef = strings.TrimSpace(typeRef)
		}
	}

	// Remove array/slice brackets: [], []*,  [n]
	for strings.HasPrefix(typeRef, "[") {
		idx := strings.Index(typeRef, "]")
		if idx == -1 {
			break
		}
		typeRef = typeRef[idx+1:]
		typeRef = strings.TrimSpace(typeRef)
	}

	// Remove pointer prefix
	typeRef = strings.TrimPrefix(typeRef, "*")
	typeRef = strings.TrimSpace(typeRef)

	// Take only the last part if it contains a dot (qualified name)
	if strings.Contains(typeRef, ".") {
		parts := strings.Split(typeRef, ".")
		typeRef = parts[len(parts)-1]
	}

	typeRef = strings.TrimSpace(typeRef)

	// Skip built-in types
	if isBuiltinType(typeRef) {
		return ""
	}

	return typeRef
}

// isBuiltinType checks if a type is a Go built-in type.
func isBuiltinType(t string) bool {
	builtins := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"float32": true, "float64": true, "complex64": true, "complex128": true,
		"bool": true, "byte": true, "rune": true, "error": true,
		"any": true, "comparable": true,
	}
	return builtins[t]
}

// findEdgeByName finds an edge in the list by symbol name (within same package).
func findEdgeByName(edges []*Edge, name string) *Edge {
	for _, e := range edges {
		if e.SymbolName == name && e.Receiver == "" {
			return e
		}
	}
	return nil
}

// findEdgeInPackageOrExternal finds an edge by type name, checking local package first.
func findEdgeInPackageOrExternal(edges []*Edge, importPath, typeName string) *Edge {
	// First try to find in local edges
	for _, e := range edges {
		if e.SymbolName == typeName && e.Receiver == "" {
			return e
		}
	}

	// If not found locally, create placeholder for external type
	// This allows tracking relationships to external types
	return &Edge{
		ImportPath: "external",
		SymbolName: typeName,
		SymbolKind: TypeKind,
		IsExported: true,
	}
}

// inferTestTarget extracts the target symbol name from a test function name.
// Examples:
//
//	"TestMyType" -> "MyType"
//	"TestMyFunc_Case1" -> "MyFunc"
//	"Test_myFunc" -> "" (don't infer for unexported targets)
func inferTestTarget(testFuncName string) string {
	if !strings.HasPrefix(testFuncName, "Test") {
		return ""
	}

	name := strings.TrimPrefix(testFuncName, "Test")
	if name == "" {
		return ""
	}

	// Handle Test_name format (underscore separator)
	if strings.HasPrefix(name, "_") {
		name = strings.TrimPrefix(name, "_")
	}

	// Extract only the part before underscore (subtest indicator)
	if idx := strings.Index(name, "_"); idx != -1 {
		name = name[:idx]
	}

	// Skip if target is not exported (starts with lowercase)
	if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
		return ""
	}

	return name
}
