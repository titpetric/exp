package jsonschema

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/titpetric/exp/cmd/schema-gen/extract"
	"github.com/titpetric/exp/cmd/schema-gen/model"
)

// ParseAndConvertStruct parses the given repo directory for Go structs and
// converts the specified rootType to JSON Schema, writing the result to "schema.json".

func ParseAndConvertStruct(cfg *options) error {
	if cfg.outputFile == "" {
		cfg.outputFile = "schema.json"
	}

	absDir, err := normalizeSourcePath(cfg.sourcePath)
	if err != nil {
		return err
	}

	pkgInfos, err := extract.Extract(absDir, &model.ExtractOptions{IncludeInternal: cfg.includeInternal})
	if err != nil {
		return err
	}
	if len(pkgInfos) == 0 {
		return fmt.Errorf("no package info extracted from %q", absDir)
	}

	schema, err := ConvertToJSONSchema(pkgInfos[0], NewDefaultConfig(), cfg)

	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(schema, "", "    ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.outputFile, jsonBytes, 0o644); err != nil {
		return err
	}

	return nil
}

// ConvertToJSONSchema converts PackageInfo to JSON Schema with only the root type and its (internal and external) dependencies.
func ConvertToJSONSchema(pkgInfo *model.PackageInfo, config *RequiredFieldsConfig, cfg *options) (*model.JSONSchema, error) {
	rootSchema := &model.JSONSchema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Definitions: make(map[string]*model.JSONSchema),
	}
	definitions := rootSchema.Definitions

	// We'll store discovered dependencies in this map
	dependencies := make(map[string]bool)
	// Build an alias mapping from the root package's imports
	aliasMap := buildAliasMap(pkgInfo.Imports)

	// Find the root type and collect its dependencies
	var rootTypeInfo *model.TypeInfo
	for _, decl := range pkgInfo.Declarations {
		for _, typ := range decl.Types {
			if typ.Name == cfg.rootType {
				rootTypeInfo = typ
				CollectDependencies(typ, pkgInfo, dependencies)
				break
			}
		}
	}
	if rootTypeInfo == nil {
		return nil, fmt.Errorf("root type %q not found in package", cfg.rootType)
	}

	// Process internal types (no dot in their name) to generate JSON Schema definitions
	for _, decl := range pkgInfo.Declarations {
		for _, typ := range decl.Types {
			// If the type is either the rootType or a discovered dependency
			if typ.Name == cfg.rootType || dependencies[typ.Name] {
				// Only handle if it's an internal type (no dot in the name)
				if !strings.Contains(typ.Name, ".") {
					schema := generateTypeSchema(typ, config, "", cfg.stripPrefix)
					if schema != nil {
						// Store it in the definitions map
						definitions[typ.Name] = schema
					}
				}
			}
		}
	}

	// Process external dependencies recursively
	visited := make(map[string]bool)
	for dep := range dependencies {
		if strings.Contains(dep, ".") {
			if err := ProcessExternalType(dep, aliasMap, definitions, visited, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			}
		}
	}

	rootSchema.Ref = "#/definitions/" + cfg.rootType
	return rootSchema, nil
}

// ProcessExternalType loads an external package for a qualified type (e.g. "model.Inner"),
// generates its JSON Schema definition, and then recursively processes its custom fields.
func ProcessExternalType(qualifiedType string, aliasMap map[string]string, definitions map[string]*model.JSONSchema, visited map[string]bool, cfg *options) error {
	if visited[qualifiedType] {
		return nil
	}
	visited[qualifiedType] = true

	parts := strings.SplitN(qualifiedType, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid qualified type: %s", qualifiedType)
	}
	pkgAlias := parts[0]
	typeName := parts[1]
	pkgPath, ok := aliasMap[pkgAlias]
	if !ok {
		return fmt.Errorf("alias %q not found in alias map", pkgAlias)
	}
	absDir, err := normalizeSourcePath(cfg.sourcePath)
	if err != nil {
		return err
	}
	extPkgInfo, err := LoadExternalPackage(pkgPath, absDir, cfg.includeInternal)
	if err != nil {
		return fmt.Errorf("failed to load external package %q: %w", pkgPath, err)
	}
	extAliasMap := buildAliasMap(extPkgInfo.Imports)
	extAliasMap[extPkgInfo.Name] = pkgPath
	var extType *model.TypeInfo
	for _, decl := range extPkgInfo.Declarations {
		for _, t := range decl.Types {
			if t.Name == typeName {
				extType = t
				if t.Type != "" && t.Name != t.Type && isCustomType(t.Type) {
					depQualified := qualifyTypeName(t.Type, pkgAlias)
					if shouldAddPreviousImports(t.Type, pkgAlias, extAliasMap) {
						if value, exists := aliasMap[pkgAlias]; exists {
							extAliasMap[pkgAlias] = value
						}
					}
					if err := ProcessExternalType(depQualified, extAliasMap, definitions, visited, cfg); err != nil {
						return err
					}
				}
				break
			}
		}
	}
	if extType == nil {
		return fmt.Errorf("type %q not found in external package %q", typeName, pkgPath)
	}
	extSchema := generateTypeSchema(extType, &RequiredFieldsConfig{Fields: map[string][]string{}}, pkgAlias, cfg.stripPrefix)
	if extSchema != nil {
		definitions[getRefName(qualifiedType, pkgAlias, cfg.stripPrefix)] = extSchema
	}
	for _, field := range extType.Fields {
		baseType := getBaseType(field.Type)
		if isCustomType(baseType) {
			if shouldAddPreviousImports(baseType, pkgAlias, extAliasMap) {
				if value, exists := aliasMap[pkgAlias]; exists {
					extAliasMap[pkgAlias] = value
				}
			}
			depQualified := qualifyTypeName(baseType, pkgAlias)
			if err := ProcessExternalType(depQualified, extAliasMap, definitions, visited, cfg); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadExternalPackage uses golang.org/x/tools/go/packages to load a package from its import path
// and then runs the extraction process on it.
func LoadExternalPackage(pkgPath, repoDir string, includeInternal bool) (*model.PackageInfo, error) {
	absDir, err := filepath.Abs(repoDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		Dir:  absDir,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no package found for %q", pkgPath)
	}
	if len(pkgs[0].GoFiles) == 0 {
		return nil, fmt.Errorf("external package %q has no Go files", pkgPath)
	}

	pkgDir := filepath.Dir(pkgs[0].GoFiles[0])
	pkgInfos, err := extract.Extract(pkgDir+"/", &model.ExtractOptions{IncludeInternal: includeInternal})
	if err != nil {
		return nil, err
	}
	if len(pkgInfos) == 0 {
		return nil, fmt.Errorf("no package info extracted for %q", pkgPath)
	}
	return pkgInfos[0], nil
}

// CollectDependencies recursively collects type dependencies from a given struct type's fields.
// If a field references a named custom type, we also parse that named type's definition
// to discover deeper dependencies (like "CertData" inside "CertsData").
func CollectDependencies(typeInfo *model.TypeInfo, pkgInfo *model.PackageInfo, dependencies map[string]bool) {
	for _, field := range typeInfo.Fields {
		if strings.HasPrefix(field.Type, "map[") {
			handleMapField(field.Type, pkgInfo, dependencies)
			continue
		}
		baseType := getBaseType(field.Type)
		if isCustomType(baseType) {
			if !dependencies[baseType] {
				dependencies[baseType] = true
				if !strings.Contains(baseType, ".") {
					// e.g. baseType == "CertsData"
					for _, decl := range pkgInfo.Declarations {
						for _, depType := range decl.Types {
							if depType.Name == baseType {
								if depType.Type != "" && depType.Name != depType.Type {
									dependencies[depType.Type] = true
								}
								// This named type might be "[]CertData", "map[string]PortWhiteList", etc.
								CollectTypeDefinitionDeps(depType, pkgInfo, dependencies)
							}
						}
					}
				}
			}
		}
	}
}

// CollectTypeDefinitionDeps inspects a named type's underlying type
// (e.g. "CertsData" -> "[]CertData") to find further dependencies.
func CollectTypeDefinitionDeps(typeInfo *model.TypeInfo, pkgInfo *model.PackageInfo, dependencies map[string]bool) {
	underlying := typeInfo.Type
	// If it's a slice: e.g. "[]CertData"
	if strings.HasPrefix(underlying, "[]") {
		elemType := strings.TrimPrefix(underlying, "[]")
		elemType = strings.TrimPrefix(elemType, "*")
		if isCustomType(elemType) {
			if !dependencies[elemType] {
				dependencies[elemType] = true
				if !strings.Contains(elemType, ".") {
					for _, decl := range pkgInfo.Declarations {
						for _, depType := range decl.Types {
							if depType.Name == elemType {
								if depType.Type != "" && depType.Name != depType.Type {
									dependencies[depType.Type] = true
								}
								CollectTypeDefinitionDeps(depType, pkgInfo, dependencies)
							}
						}
					}
				}
			}
		}
		return
	}

	// If it's a map: e.g. "map[string]PortWhiteList"
	if strings.HasPrefix(underlying, "map[") {
		handleMapField(underlying, pkgInfo, dependencies)
		return
	}

	if len(typeInfo.Fields) > 0 {
		for _, field := range typeInfo.Fields {
			if strings.HasPrefix(field.Type, "map[") {
				handleMapField(field.Type, pkgInfo, dependencies)
				continue
			}
			baseType := getBaseType(field.Type)
			if isCustomType(baseType) {
				if !dependencies[baseType] {
					dependencies[baseType] = true
					if !strings.Contains(baseType, ".") {
						for _, decl := range pkgInfo.Declarations {
							for _, depType := range decl.Types {
								if depType.Name == baseType {
									if depType.Type != "" && depType.Name != depType.Type {
										dependencies[depType.Type] = true
									}
									CollectTypeDefinitionDeps(depType, pkgInfo, dependencies)
								}
							}
						}
					}
				}
			}
		}
		return
	}

}

// GenerateEnumSchema creates a JSON Schema definition for an enum type.
func GenerateEnumSchema(typeInfo *model.TypeInfo) *model.JSONSchema {
	enumValues := make([]any, 0, len(typeInfo.Enums))
	for _, enum := range typeInfo.Enums {
		enumValues = append(enumValues, enum.Value)
	}
	jsonType := "string"
	if typeInfo.Type == "int" {
		jsonType = "integer"
	}
	return &model.JSONSchema{
		Type: jsonType,
		Enum: enumValues,
	}
}

// GenerateStructSchema creates a JSON Schema definition for a struct type.
func GenerateStructSchema(typeInfo *model.TypeInfo, config *RequiredFieldsConfig, pkgName string, stripPrefix []string) *model.JSONSchema {
	schema := &model.JSONSchema{
		Type:                 "object",
		Properties:           make(map[string]*model.JSONSchema),
		AdditionalProperties: false,
	}
	requiredFields := config.Fields[typeInfo.Name]
	requiredMap := make(map[string]bool)
	for _, field := range requiredFields {
		requiredMap[field] = true
	}
	var required []string

	for _, field := range typeInfo.Fields {

		if field.JSONName == "-" || field.JSONName == "" {
			continue
		}
		isArray := strings.HasPrefix(field.Type, "[]")
		baseType := getBaseType(field.Type)
		var fieldSchema *model.JSONSchema
		if isCustomType(baseType) {
			refName := getRefName(baseType, pkgName, stripPrefix)
			if isArray {
				fieldSchema = &model.JSONSchema{
					Type: "array",
					Items: &model.JSONSchema{
						Ref: "#/definitions/" + refName,
					},
				}
			} else {
				fieldSchema = &model.JSONSchema{
					Ref: "#/definitions/" + refName,
				}
			}
		} else {
			fieldSchema = getJSONType(field.Type)
		}
		if field.Doc != "" {
			fieldSchema.Description = field.Doc
		}
		cleanedJson := parseJSONTag(field.JSONName)
		schema.Properties[cleanedJson] = fieldSchema
		if requiredMap[field.Name] {
			required = append(required, cleanedJson)
		}
	}
	if len(required) > 0 {
		schema.Required = required
	}
	return schema
}

// GenerateMapDefinition creates a top-level JSON Schema definition for a map type (e.g. map[string]Something).
func GenerateMapDefinition(goType string) *model.JSONSchema {
	// Example: "map[string]interface{}" or "map[string]PortWhiteList"
	inside := goType[len("map["):]
	parts := strings.SplitN(inside, "]", 2)
	if len(parts) != 2 {
		return &model.JSONSchema{
			Type:                 "object",
			AdditionalProperties: true,
		}
	}
	keyType := strings.TrimSpace(parts[0])   // e.g. "string"
	valueType := strings.TrimSpace(parts[1]) // e.g. "interface{}" or "PortWhiteList"

	if keyType != "string" {
		return &model.JSONSchema{
			Type:                 "object",
			AdditionalProperties: true,
		}
	}

	if valueType == "interface{}" || valueType == "any" {
		return &model.JSONSchema{
			Type:                 "object",
			AdditionalProperties: true,
		}
	}

	if !isCustomType(valueType) {
		return &model.JSONSchema{
			Type: "object",
			AdditionalProperties: &model.JSONSchema{
				Type: getBaseJSONType(valueType),
			},
		}
	}

	return &model.JSONSchema{
		Type: "object",
		AdditionalProperties: &model.JSONSchema{
			Ref: "#/definitions/" + valueType,
		},
	}
}

// GenerateSliceDefinition creates a top-level JSON Schema definition for a slice type (e.g. []CertData).
func GenerateSliceDefinition(goType string) *model.JSONSchema {
	elemType := strings.TrimPrefix(goType, "[]")
	elemType = strings.TrimSpace(elemType)

	if !isCustomType(elemType) {
		return &model.JSONSchema{
			Type: "array",
			Items: &model.JSONSchema{
				Type: getBaseJSONType(elemType),
			},
		}
	}

	return &model.JSONSchema{
		Type: "array",
		Items: &model.JSONSchema{
			Ref: "#/definitions/" + elemType,
		},
	}
}

// RequiredFieldsConfig defines which fields are required for each type.
type RequiredFieldsConfig struct {
	Fields map[string][]string
}

// NewDefaultConfig just returns a sample required-fields config
func NewDefaultConfig() *RequiredFieldsConfig {
	return &RequiredFieldsConfig{
		Fields: map[string][]string{
			//"User":  {"ID", "Name"}, // Only ID and Name are required for User
			//"Inner": {"Name"},
		},
	}
}

func generateTypeSchema(typ *model.TypeInfo, config *RequiredFieldsConfig, pkgName string, stripPrefix []string) *model.JSONSchema {
	switch {
	case len(typ.Enums) > 0:
		return GenerateEnumSchema(typ)

	case len(typ.Fields) > 0:
		return GenerateStructSchema(typ, config, pkgName, stripPrefix)

	case strings.HasPrefix(typ.Type, "map["):
		return GenerateMapDefinition(typ.Type)

	case strings.HasPrefix(typ.Type, "[]"):
		return GenerateSliceDefinition(typ.Type)

	case !isCustomType(typ.Type):
		return &model.JSONSchema{Type: getBaseJSONType(typ.Type)}

	case typ.Name != typ.Type && len(typ.Fields) == 0:
		refName := getRefName(typ.Type, pkgName, stripPrefix)
		return &model.JSONSchema{Ref: "#/definitions/" + refName}

	default:
		log.Printf("Skipping type %q with underlying type %q\n", typ.Name, typ.Type)
		return nil
	}
}
