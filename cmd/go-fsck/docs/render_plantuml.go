package docs

import (
	"fmt"
	gast "go/ast"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/internal/ast"
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// dbRelationship describes a database relationship
type dbRelationship struct {
	fromType   string
	fromField  string
	toType     string
	isOneToOne bool
	isExternal bool // true if toType is not defined in the codebase (conceptual type)
}

// detectDBRelationships scans a type for database relationships.
// It looks for fields with "ID" suffix and checks if the target type has an ID field.
// If the target type has an ID field, it's a 1:N relationship (fromType owns many toType).
// If the target type doesn't have an ID field, it's a 1:1 relationship (the ID field is the PK).
// Unresolved foreign keys (no matching type) are marked as isExternal=true (conceptual types).
func detectDBRelationships(t *model.Declaration, allTypes map[string]*model.Declaration) []dbRelationship {
	var relations []dbRelationship

	for _, f := range t.Fields {
		// Check if field name ends with "ID" or "Id" (case-insensitive)
		// Examples: UserID, PostId, post_id, etc.
		var baseName string
		fieldNameLower := strings.ToLower(f.Name)

		if strings.HasSuffix(fieldNameLower, "id") {
			// For PascalCase like "UserID" or "PostId"
			if len(f.Name) > 2 && (f.Name[len(f.Name)-2:] == "ID" || f.Name[len(f.Name)-2:] == "Id") {
				baseName = f.Name[:len(f.Name)-2]
			} else if strings.HasSuffix(fieldNameLower, "_id") {
				// For snake_case like "user_id" or "post_id"
				baseName = strings.TrimSuffix(fieldNameLower, "_id")
			} else {
				continue
			}
		} else {
			continue
		}

		// Try to find the target type by checking database column name from tag
		var targetTypeName string
		var exists bool

		if f.Tag != "" {
			dbCol := ast.DBTagName(f.Tag)
			if dbCol != "" && strings.HasSuffix(strings.ToLower(dbCol), "_id") {
				// Extract target type from column name with fallback matching
				baseCol := strings.TrimSuffix(strings.ToLower(dbCol), "_id")
				targetTypeName, exists = tryTypeMatches(baseCol, allTypes)
			}
		}

		// If we couldn't find it from the tag, try the field name with fallback matching
		if targetTypeName == "" {
			targetTypeName, exists = tryTypeMatches(baseName, allTypes)
		}

		// If target type was not found, use the base name as a conceptual type (interface)
		// This handles cases like OwnerID -> Owner, or ActorID -> Actor (if they don't exist)
		isExternal := false
		if !exists {
			targetTypeName = toPascalCase(baseName)
			isExternal = true
		}

		// Determine if it's 1:1 or 1:N by checking if source type has its own ID field.
		// If source has ID, it can have multiple records with the same FK → 1:N
		// If source doesn't have ID, the FK is its only key → 1:1 (unique FK)
		sourceHasID := hasIDField(t)
		isOneToOne := !sourceHasID

		relations = append(relations, dbRelationship{
			fromType:   t.Name,
			fromField:  f.Name,
			toType:     targetTypeName,
			isOneToOne: isOneToOne,
			isExternal: isExternal,
		})
	}

	return relations
}

// hasIDField checks if a type has an "ID" or "Id" field (the primary key).
func hasIDField(t *model.Declaration) bool {
	for _, f := range t.Fields {
		if f.Name == "ID" || f.Name == "Id" {
			return true
		}
	}
	return false
}

// toPascalCase converts snake_case to PascalCase.
// e.g., "user_profile" -> "UserProfile", "user" -> "User"
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// tryTypeMatches tries to find a matching type, with fallback logic for prefixed field names.
// For example, "actor_user" will try "ActorUser" first, then "User".
// This handles cases like actor_user_id -> User, member_user_id -> User, etc.
func tryTypeMatches(baseName string, allTypes map[string]*model.Declaration) (string, bool) {
	// Try direct match first
	candidate := toPascalCase(baseName)
	if _, exists := allTypes[candidate]; exists {
		return candidate, true
	}

	// For snake_case with multiple parts, try progressively shorter suffixes by removing prefixes
	// This handles: actor_user -> try "ActorUser", then "User"
	parts := strings.Split(baseName, "_")
	for i := len(parts) - 1; i >= 1; i-- {
		candidate := toPascalCase(strings.Join(parts[i:], "_"))
		if _, exists := allTypes[candidate]; exists {
			return candidate, true
		}
	}

	return "", false
}

// parseHideList parses a comma-separated list of type names to hide
func parseHideList(hideStr string) map[string]bool {
	hideMap := make(map[string]bool)
	if hideStr == "" {
		return hideMap
	}
	parts := strings.Split(hideStr, ",")
	for _, part := range parts {
		hideMap[strings.TrimSpace(part)] = true
	}
	return hideMap
}

func renderPlantUML(opt *options, defs []*model.Definition) error {
	var links []string
	externalTypes := make(map[string]bool) // Track conceptual/external types

	addLink := func(link string) {
		links = append(links, link)
	}

	// Parse hide list
	hideMap := parseHideList(opt.hide)

	fmt.Println("@startuml")
	fmt.Println("")

	allTypes := make(map[string]*model.Declaration)
	allPackages := make(map[string]*model.Package)
	allFuncs := make(map[string][]*model.Declaration)

	for _, def := range defs {
		allPackages[def.Package.ImportPath] = &def.Package
		for _, t := range def.Types {
			allTypes[t.Name] = t
		}
	}

	for _, def := range defs {
		for _, t := range def.Funcs {
			receiver := t.ReceiverTypeRef()
			if receiver == "" {
				continue
			}

			receiver = def.Package.Namespace(".") + receiver

			allFuncs[receiver] = append(allFuncs[receiver], t)
		}
	}

	for _, def := range defs {
		importMap, _ := def.Imports.Map(def.Imports.All())

		lookup := func(name string) (*model.Package, bool) {
			importpath, ok := importMap[name]
			if !ok {
				return nil, false
			}

			pkg, ok := allPackages[importpath]
			return pkg, ok
		}

		if def.Package.TestPackage || strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}

		namespace := def.Package.Namespace(".")

		for _, t := range def.Types {
			if len(t.Fields) == 0 {
				continue
			}

			// Skip hidden types
			if hideMap[t.Name] {
				continue
			}

			// Skip interfaces in model mode
			if opt.model && strings.HasPrefix(t.Type, "interface") {
				continue
			}

			// Detect database relationships
			dbRelations := detectDBRelationships(t, allTypes)
			for _, rel := range dbRelations {
				relType := "1:N"
				if rel.isOneToOne {
					relType = "1:1"
				}
				// Track external types for later rendering
				if rel.isExternal {
					externalTypes[rel.toType] = true
				}
				addLink(fmt.Sprintf("%q --> %q : .%s (%s)", namespace+t.Name, namespace+rel.toType, rel.fromField, relType))
			}

			for _, name := range t.GetNames() {
				var token = "class"
				if strings.HasPrefix(t.Type, "interface") {
					token = "interface"
				}

				for _, f := range t.Fields {
					if strings.HasPrefix(f.Type, "func") {
						token = "interface"
					}
				}

				if len(t.Arguments) > 0 {
					names := []string{}
					for _, name := range t.Arguments {
						names = append(names, name)
					}
					t.Arguments = names
					name += "[" + strings.Join(names, ", ") + "]"
				}

				fmt.Println(token, fmt.Sprintf("%q", namespace+name), "{")
				for _, f := range t.Fields {
					typeRef := f.TypeRef()

					if f.Name == "" {
						if strings.Contains(typeRef, ".") {
							parts := strings.SplitN(typeRef, ".", 2)
							packageName, typeName := parts[0], parts[1]
							if p, ok := lookup(packageName); ok {
								addLink(fmt.Sprintf("%q --|> %q : embeds", namespace+name, p.Namespace(".")+typeName))
							}
						} else {
							addLink(fmt.Sprintf("%q --|> %q : embeds", namespace+name, namespace+typeRef))
						}
						continue
					}

					if strings.HasPrefix(f.Type, "struct") {
						f.Type = "struct"
					}
					if strings.HasPrefix(f.Type, "interface") {
						f.Type = "interface"
					}

					if gast.IsExported(f.Name) {
						fmt.Println("  +", f.Name+":", f.Type)
					} else {
						fmt.Println("  -", f.Name+":", f.Type)
					}

					if token != "interface" {
						if strings.Contains(typeRef, ".") {
							parts := strings.SplitN(typeRef, ".", 2)
							packageName, typeName := parts[0], parts[1]
							if p, ok := lookup(packageName); ok {
								addLink(fmt.Sprintf("%q --> %q : .%s", namespace+name, p.Namespace(".")+typeName, f.Name))
							}
						} else {
							if _, ok := model.ToType(typeRef); ok {
								addLink(fmt.Sprintf("%q --> %q : .%s", namespace+name, namespace+typeRef, f.Name))
							}
						}
					}
				}
				if opt.docs && t.Doc != "" {
					addLink("")
					addLink("note top of " + namespace + name)
					addLink(t.Doc)
					addLink("end note")
					addLink("")
				}

				addLink("")

				if token == "interface" {
					continue
				}

				// Skip functions in model mode
				if opt.model {
					continue
				}

				funcList := allFuncs[namespace+t.Name]
				for _, sig := range funcList {
					funcInfo := sig
					funcName := sig.Name

					func() {
						for _, argType := range funcInfo.Returns {
							typeRef := model.TypeRef(argType)

							if strings.Contains(typeRef, ".") {
								parts := strings.SplitN(typeRef, ".", 2)
								packageName, typeName := parts[0], parts[1]
								if p, ok := lookup(packageName); ok {
									addLink(fmt.Sprintf("%q --> %q : .%s()", namespace+name, p.Namespace(".")+typeName, funcName))
									return
								}
							}
						}
						for _, argType := range funcInfo.Arguments {
							typeRef := model.TypeRef(argType)

							if strings.Contains(typeRef, ".") {
								parts := strings.SplitN(typeRef, ".", 2)
								packageName, typeName := parts[0], parts[1]
								if p, ok := lookup(packageName); ok {
									addLink(fmt.Sprintf("%q --> %q : .%s()", namespace+name, p.Namespace(".")+typeName, funcName))
									return
								}
							}
						}
					}()

					if gast.IsExported(sig.Signature) {
						fmt.Println("  +", sig.Signature)
					} else {
						// fmt.Println("  -", sig)
					}
				}

			}

			fmt.Println("}")
			fmt.Println()
		}
	}

	// Render conceptual/external types as interfaces
	if len(externalTypes) > 0 {
		fmt.Println()
		for typeName := range externalTypes {
			fmt.Println(fmt.Sprintf("interface %q", typeName))
		}
	}

	for _, link := range links {
		fmt.Println(link)
	}
	if len(links) > 0 {
		fmt.Println()
	}

	fmt.Println("@enduml")

	return nil
}
