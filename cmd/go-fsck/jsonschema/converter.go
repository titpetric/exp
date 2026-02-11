package jsonschema

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ParseAndConvertStruct parses the given repo directory for Go structs and
// converts the specified rootType to JSON Schema, writing the result to "schema.json".
func ParseAndConvertStruct(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	schema, err := ConvertToJSONSchema(defs[0], NewDefaultConfig(), cfg)
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
func ConvertToJSONSchema(pkgInfo *model.Definition, config *RequiredFieldsConfig, cfg *options) (*model.JSONSchema, error) {
	rootSchema := &model.JSONSchema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Definitions: make(map[string]*model.JSONSchema),
	}
	definitions := rootSchema.Definitions

	for _, decl := range pkgInfo.Types {
		if decl.Type == "interface" {
			continue
		}
		schema := generateTypeSchema(decl, config, "", cfg.stripPrefix)
		if schema != nil {
			if decl.Doc != "" {
				schema.Description = decl.Doc
			}
			definitions[decl.Name] = schema
		}
	}
	// rootSchema.Ref = "#/definitions/" + cfg.rootType
	return rootSchema, nil
}

func generateTypeSchema(decl *model.Declaration, config *RequiredFieldsConfig, pkgName string, stripPrefix []string) *model.JSONSchema {
	switch {
	case len(decl.Fields) > 0:
		return GenerateStructSchema(decl, config, pkgName, stripPrefix)

	case strings.HasPrefix(decl.Type, "map["):
		return GenerateMapDefinition(decl.Type)

	case strings.HasPrefix(decl.Type, "[]"):
		return GenerateSliceDefinition(decl.Type)

	case !isCustomType(decl.Type):
		return &model.JSONSchema{Type: getBaseJSONType(decl.Type)}

	case strings.Contains(decl.Type, "."):
		log.Printf("Skipping %q with external ref %q", decl.Name, decl.Type)
		return nil

	case decl.Name != decl.Type && len(decl.Fields) == 0:
		refName := getRefName(decl.Type, pkgName, stripPrefix)
		return &model.JSONSchema{Ref: "#/definitions/" + refName}

	default:
		log.Printf("Skipping %q with underlying type %q\n", decl.Name, decl.Type)
		return nil
	}
}

func generateFieldSchema(decl *model.Field, config *RequiredFieldsConfig, pkgName string, stripPrefix []string) *model.JSONSchema {
	switch {
	case strings.HasPrefix(decl.Type, "map["):
		return GenerateMapDefinition(decl.Type)

	case strings.HasPrefix(decl.Type, "[]"):
		return GenerateSliceDefinition(decl.Type)

	case !isCustomType(decl.Type):
		return &model.JSONSchema{Type: getBaseJSONType(decl.Type)}

	default:
		log.Printf("Skipping %q with underlying type %q\n", decl.Name, decl.Type)
		return nil
	}
}

// CollectTypeDefinitionDeps inspects a named type's underlying type
// (e.g. "CertsData" -> "[]CertData") to find further dependencies.
func CollectTypeDefinitionDeps(typeInfo *model.Declaration, pkgInfo *model.Definition, dependencies map[string]bool) {
	underlying := typeInfo.Type
	// If it's a slice: e.g. "[]CertData"
	if strings.HasPrefix(underlying, "[]") {
		elemType := strings.TrimPrefix(underlying, "[]")
		elemType = strings.TrimPrefix(elemType, "*")
		if isCustomType(elemType) {
			if !dependencies[elemType] {
				dependencies[elemType] = true
				if !strings.Contains(elemType, ".") {
					for _, depType := range pkgInfo.Types {
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
						for _, depType := range pkgInfo.Types {
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
		return
	}
}

// GenerateStructSchema creates a JSON Schema definition for a struct type.
func GenerateStructSchema(typeInfo *model.Declaration, config *RequiredFieldsConfig, pkgName string, stripPrefix []string) *model.JSONSchema {
	result := &model.JSONSchema{
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

		schema := generateFieldSchema(field, config, pkgName, stripPrefix)
		if schema == nil {
			continue
		}

		if field.Doc != "" {
			schema.Description = field.Doc
		}
		cleanedJson := parseJSONTag(field.JSONName)

		result.Properties[cleanedJson] = schema

		if requiredMap[field.Name] {
			required = append(required, cleanedJson)
		}
	}
	if len(required) > 0 {
		result.Required = required
	}
	return result
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
