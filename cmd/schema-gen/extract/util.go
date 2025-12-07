package extract

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"
)

func TrimSpace(in *ast.CommentGroup) string {
	if in != nil {
		return strings.TrimSpace(in.Text())
	}
	return ""
}

var TypeName = getTypeDeclarations

func getTypeDeclaration(ident *ast.Ident) string {
	return ident.Name
}

func getTypeDeclarations(expr *ast.TypeSpec) string {
	return getTypeDeclarationsForExpr(expr.Type)
}

func getTypeDeclarationsForPointerType(starExpr *ast.StarExpr) string {
	return "*" + getTypeDeclarationsForExpr(starExpr.X)
}

func getTypeDeclarationsForArrayType(arrayType *ast.ArrayType) string {
	var declaration string
	switch obj := arrayType.Elt.(type) {
	case *ast.Ident:
		declaration = getTypeDeclaration(obj)
	case *ast.ArrayType:
		declaration = getTypeDeclarationsForArrayType(obj)
	case *ast.MapType:
		declaration = getTypeDeclarationsForMapType(obj)
	case *ast.StarExpr:
		declaration = getTypeDeclarationsForPointerType(obj)
	case *ast.SelectorExpr:
		declaration = getTypeDeclarationsForExpr(obj.X) + "." + getTypeDeclaration(obj.Sel)
	default:
		declaration = fmt.Sprintf("%#v", arrayType.Elt)
	}

	return fmt.Sprintf("[]%s", declaration)
}

func getTypeDeclarationsForMapType(mapType *ast.MapType) string {
	keyIdent, ok := mapType.Key.(*ast.Ident)
	if !ok {
		fmt.Fprintln(os.Stderr, "WARN: unsupported key type in map")
		return "any"
	}
	keyType := getTypeDeclaration(keyIdent)

	valueIdent, ok := mapType.Value.(*ast.Ident)
	if !ok {
		if pointerType, isPointerType := mapType.Value.(*ast.StarExpr); isPointerType {
			// If Value is a pointer type, get the underlying type name
			valueType := getTypeDeclaration(pointerType.X.(*ast.Ident))
			return fmt.Sprintf("map[%s]%s", keyType, valueType)
		}
		return fmt.Sprintf("map[%s]interface{}", keyType)
	}
	valueType := getTypeDeclaration(valueIdent)

	return fmt.Sprintf("map[%s]%s", keyType, valueType)
}

func getTypeDeclarationsForExpr(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.SelectorExpr:
		return getTypeDeclarationsForExpr(expr.X) + "." + getTypeDeclaration(expr.Sel)
	case *ast.Ident:
		return getTypeDeclaration(expr)
	case *ast.ArrayType:
		return getTypeDeclarationsForArrayType(expr)
	case *ast.MapType:
		return getTypeDeclarationsForMapType(expr)
	case *ast.InterfaceType:
		return ""
	case *ast.StructType:
		return "struct{}"
	case *ast.StarExpr:
		return getTypeDeclarationsForExpr(expr.X)
	case *ast.FuncType:
		return "func"
	default:
		return fmt.Sprintf("%#v", expr)
	}
}

func jsonTag(tag string) string {
	return reflect.StructTag(tag).Get("json")
}

func write(cfg *options) error {
	opts := NewExtractOptions(cfg)

	definitions, err := Extract(cfg.sourcePath, opts)
	if err != nil {
		return err
	}

	output := os.Stdout
	switch cfg.outputFile {
	case "", "-":
	default:
		var err error
		output, err = os.Create(cfg.outputFile)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(output)
	if cfg.prettyJSON {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(definitions)
}
