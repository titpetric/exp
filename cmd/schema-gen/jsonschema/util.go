package jsonschema

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/titpetric/exp/cmd/schema-gen/model"
)

func Title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func getRefName(baseType, pkgName string, stripPrefix []string) string {
	pkg, refType := getPkgAndBaseType(baseType, pkgName)
	var refName string
	if slices.Contains(stripPrefix, pkg) || pkg == "" {
		refName = refType
	} else {

		refName = Title(pkg) + refType
	}
	return refName
}
func handleMapField(fieldType string, pkgInfo *model.PackageInfo, dependencies map[string]bool) {
	// e.g. "map[string]RequestHeadersRewriteConfig"
	inside := fieldType[len("map["):]
	parts := strings.SplitN(inside, "]", 2)
	if len(parts) != 2 {
		return
	}
	valueType := strings.TrimSpace(parts[1])
	valueType = strings.TrimPrefix(valueType, "*")

	if isCustomType(valueType) {
		if !dependencies[valueType] {
			dependencies[valueType] = true
			if !strings.Contains(valueType, ".") {
				// find its definition, parse deeper
				for _, decl := range pkgInfo.Declarations {
					for _, depType := range decl.Types {
						if depType.Name == valueType {
							CollectTypeDefinitionDeps(depType, pkgInfo, dependencies)
						}
					}
				}
			}
		}
	}
}

func getJSONType(goType string) *model.JSONSchema {
	if goType == "[]byte" {
		return &model.JSONSchema{
			Type:   "string",
			Format: "byte",
		}
	}

	if strings.HasPrefix(goType, "[]") {
		elementType := strings.TrimPrefix(goType, "[]")
		if isCustomType(elementType) {
			return &model.JSONSchema{
				Type: "array",
				Items: &model.JSONSchema{
					Ref: "#/definitions/" + elementType,
				},
			}
		}
		return &model.JSONSchema{
			Type: "array",
			Items: &model.JSONSchema{
				Type: getBaseJSONType(elementType),
			},
		}
	}

	if strings.HasPrefix(goType, "map[") {
		inside := goType[len("map["):]
		parts := strings.SplitN(inside, "]", 2)
		if len(parts) != 2 {
			return &model.JSONSchema{
				Type:                 "object",
				AdditionalProperties: true,
			}
		}
		keyType := strings.TrimSpace(parts[0])
		valueType := strings.TrimSpace(parts[1])

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
		// custom type => $ref
		return &model.JSONSchema{
			Type: "object",
			AdditionalProperties: &model.JSONSchema{
				Ref: "#/definitions/" + valueType,
			},
		}
	}

	schema := &model.JSONSchema{
		Type: getBaseJSONType(goType),
	}

	switch goType {
	case "uint8", "byte":
		schema.Minimum = ToPtr(0.0)
		schema.Maximum = ToPtr(255.0)
	case "uint16":
		schema.Minimum = ToPtr(0.0)
		schema.Maximum = ToPtr(65535.0)
	case "uint32":
		schema.Minimum = ToPtr(0.0)
		schema.Maximum = ToPtr(4294967295.0)
	case "uint64", "uint":
		schema.Minimum = ToPtr(0.0)
	case "int8":
		schema.Minimum = ToPtr(-128.0)
		schema.Maximum = ToPtr(127.0)
	case "int16":
		schema.Minimum = ToPtr(-32768.0)
		schema.Maximum = ToPtr(32767.0)
	case "int32", "rune":
		schema.Minimum = ToPtr(-2147483648.0)
		schema.Maximum = ToPtr(2147483647.0)
	case "time.Time":
		schema.Type = "string"
		schema.Format = "date-time"
	case "time.Duration":
		schema.Type = "string"
		schema.Pattern = "^[-+]?([0-9]*(\\.[0-9]*)?[a-z]+)+$"
	case "complex64", "complex128":
		return &model.JSONSchema{
			Type: "object",
			Properties: map[string]*model.JSONSchema{
				"real": {Type: "number"},
				"imag": {Type: "number"},
			},
			Required: []string{"real", "imag"},
		}
	}
	return schema
}

func getBaseJSONType(goType string) string {
	switch goType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte", "rune", "uintptr":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string", "time.Time", "time.Duration":
		return "string"
	case "interface{}", "any":
		return "object"
	case "error":
		return "string"
	default:
		return "string"
	}
}

func getBaseType(fieldType string) string {
	baseType := strings.TrimPrefix(fieldType, "[]")
	baseType = strings.TrimPrefix(baseType, "*")
	// Handle the special case of "[]*SomeType"
	if strings.HasPrefix(fieldType, "[]*") {
		baseType = strings.TrimPrefix(fieldType, "[]*")
	}
	return baseType
}

func isCustomType(typeName string) bool {
	if strings.HasPrefix(typeName, "map[") || typeName == "[]byte" {
		return false
	}
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"complex64", "complex128",
		"string", "bool", "interface{}", "any",
		"byte", "rune",
		"uintptr", "error",
		"time.Time", "time.Duration":
		return false
	}
	return true
}

func buildAliasMap(imports []string) map[string]string {
	aliasMap := make(map[string]string)
	for _, imp := range imports {
		parts := strings.Split(imp, " ")
		if len(parts) == 2 {
			alias := parts[0]
			path := strings.Trim(parts[1], "\"")
			aliasMap[alias] = path
		} else {
			// No alias; deduce one from the import path
			impPath := strings.Trim(imp, "\"")
			segs := strings.Split(impPath, "/")
			aliasMap[segs[len(segs)-1]] = impPath
		}
	}
	return aliasMap
}

func ToPtr[T any](v T) *T {
	return &v
}

func parseJSONTag(tagValue string) string {
	parts := strings.SplitN(tagValue, ",", 2)
	if len(parts) == 0 {
		return ""
	}
	name := parts[0]
	if name == "-" {
		return ""
	}
	return name
}

func convertQualifiedType(qType, pkgName string) (string, string) {
	parts := strings.SplitN(qType, ".", 2)
	if len(parts) < 2 {
		// Return the input unchanged if it isn't in the expected format.
		return pkgName, qType
	}
	pkg := parts[0]
	typ := parts[1]
	return pkg, typ
}

func getPkgAndBaseType(baseType, pkgName string) (string, string) {
	if strings.Contains(baseType, ".") {
		return convertQualifiedType(baseType, pkgName)
	}
	return pkgName, baseType
}

func shouldAddPreviousImports(baseType, pkgAlias string, aliasMap map[string]string) bool {
	_, exists := aliasMap[pkgAlias]
	return !strings.Contains(baseType, ".") && !exists
}

func qualifyTypeName(baseType, pkgAlias string) string {
	if strings.Contains(baseType, ".") {
		return baseType
	}
	if pkgAlias == "" {
		return baseType
	}
	return pkgAlias + "." + baseType
}

func normalizeSourcePath(sourcePath string) (string, error) {
	absDir, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", err
	}
	return absDir + "/", nil
}
