package diff

import (
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// declKind is the kind a symbol is reported under, taken from the list of the
// definition it was collected from rather than from the declaration itself:
// a struct type is declared as "struct" but reads as a type.
const (
	kindType  = "type"
	kindConst = "const"
	kindVar   = "var"
	kindFunc  = "func"
)

// The shapes a type is reported under. A struct is the shape the model records
// nothing for, since it describes it by its fields instead.
const (
	underlyingStruct    = "struct"
	underlyingInterface = "interface"
)

// entry is one exported declaration, as it is reported and as it is compared.
type entry struct {
	symbol Symbol

	// normalized is the signature with parameter names removed, which is what
	// two revisions of a func are compared on, and the shape a type is
	// declared with, which is what two revisions of a type are compared on.
	normalized string

	// fields are the exported fields of a struct, or the methods of an
	// interface, by name. It is nil for every other kind.
	fields map[string]Field
}

// symbols returns the declarations of defs, keyed so that the same declaration
// in two revisions of a package produces the same key.
//
// Test packages and commands are left out. Internal packages and unexported
// declarations are left out too unless asked for: neither is part of the API
// another module can depend on, and both are reported separately when a
// caller wants to see a refactor rather than a release.
func symbols(defs []*model.Definition, includeInternal, includeUnexported bool) map[string]entry {
	result := make(map[string]entry)
	for _, def := range defs {
		if def.TestPackage || def.Package.Package == "main" {
			continue
		}
		if !includeInternal && isInternal(def.ImportPath) {
			continue
		}

		lists := []struct {
			kind  string
			decls model.DeclarationList
		}{
			{kindType, def.Types},
			{kindConst, def.Consts},
			{kindVar, def.Vars},
			{kindFunc, def.Funcs},
		}
		for _, list := range lists {
			for _, decl := range list.decls {
				for _, name := range declNames(decl, includeUnexported) {
					symbol := Symbol{
						Key:      def.ImportPath + "." + name,
						Package:  def.ImportPath,
						Name:     name,
						Kind:     list.kind,
						Exported: exportedSymbol(def.ImportPath, name),
					}
					var (
						normalized string
						fields     map[string]Field
					)
					switch {
					case list.kind == kindFunc && decl.Signature != "":
						symbol.Signature = signature(decl)
						normalized = normalizeSignature(decl.Signature)
					case list.kind == kindType:
						// A func carries its whole body in Source, which is
						// no use to a reader; a type carries its shape, which
						// is the thing worth printing.
						symbol.Definition = sourceWithoutDoc(decl.Source)
						symbol.Underlying = underlying(decl)
						symbol.Fields = exportedFields(decl)

						// The shape reads as the declaration writes it, so a
						// type that changes it is reported the way a func that
						// changes its signature is.
						normalized = "type " + name + " " + symbol.Underlying
						fields = fieldIndex(symbol.Fields)
					}
					result[symbol.Key] = entry{symbol: symbol, normalized: normalized, fields: fields}
				}
			}
		}
	}
	return result
}

// declNames returns the names of a declaration, each qualified with
// its receiver type.
//
// Names are filtered one by one rather than through Declaration.IsExported,
// which reports on the whole group: an unexported name sharing a "const (...)"
// block with an exported one would otherwise come along with it.
func declNames(decl *model.Declaration, includeUnexported bool) []string {
	// A method is only reachable when its receiver type is.
	if !includeUnexported && decl.Receiver != "" && !ast.IsExported(model.TypeRef(decl.Receiver)) {
		return nil
	}

	var names []string
	for _, name := range decl.GetNames() {
		if !includeUnexported && !ast.IsExported(name) {
			continue
		}
		names = append(names, strings.Trim(decl.Receiver+"."+name, "*."))
	}
	return names
}

// underlying returns the shape a type is declared with: the type it is defined
// as, "interface" for an interface, and "struct" for a struct.
//
// The model records no type for a struct, describing it by its fields instead,
// so an empty one reads as a struct. Every other shape names itself.
func underlying(decl *model.Declaration) string {
	if decl.Type == "" {
		return underlyingStruct
	}
	return decl.Type
}

// exportedFields returns the exported fields of a struct, or the methods of an
// interface, in name order. Unexported fields are left out: they are not
// something another module can reach, so they are not a promise to keep.
func exportedFields(decl *model.Declaration) []Field {
	var fields []Field
	for _, field := range decl.Fields {
		name, embedded := fieldName(field)
		if !ast.IsExported(name) {
			continue
		}
		fields = append(fields, Field{
			Name:     name,
			Type:     field.Type,
			Tag:      field.Tag,
			Embedded: embedded,
		})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

// fieldName returns the name a field is reached by, and whether it is embedded.
//
// An embedded field is recorded with no name of its own, and is reached by the
// last identifier of the type it embeds: "platform.Module" is reached as
// "Module". An embedded exported type is part of the API, since it promotes its
// method set into the one embedding it.
func fieldName(field *model.Field) (name string, embedded bool) {
	if field.Name != "" {
		return field.Name, false
	}

	name = strings.TrimPrefix(field.Type, "*")
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		name = name[dot+1:]
	}
	return name, true
}

// fieldIndex keys a field list by name, which is what two revisions of a type
// are compared through.
func fieldIndex(fields []Field) map[string]Field {
	index := make(map[string]Field, len(fields))
	for _, field := range fields {
		index[field.Name] = field
	}
	return index
}

// signature renders a func the way it is declared, which is what a reader
// wants to see listed. The receiver keeps the pointer the source gave it, so
// it reads as the method set does.
func signature(decl *model.Declaration) string {
	if decl.Receiver != "" {
		return "func (" + decl.Receiver + ") " + decl.Signature
	}
	return "func " + decl.Signature
}

// sourceWithoutDoc returns a declaration as it is written, with the doc comment
// lines above it removed. Only the run of comments the declaration opens with
// is dropped, so a comment inside the body stays where it is.
//
// The model holds no source at all unless it was extracted with
// --include-sources, in which case there is nothing to return.
func sourceWithoutDoc(source string) string {
	source = strings.TrimSpace(source)
	for strings.HasPrefix(source, "//") {
		_, rest, ok := strings.Cut(source, "\n")
		if !ok {
			return ""
		}
		source = strings.TrimSpace(rest)
	}
	return source
}

// isInternal reports whether an import path holds an internal path segment,
// which keeps the package out of reach of any module but its own.
func isInternal(importPath string) bool {
	for _, segment := range strings.Split(importPath, "/") {
		if segment == "internal" {
			return true
		}
	}
	return false
}

// normalizeSignature drops the parameter names of a signature, so that renaming
// a parameter does not read as an API change while changing its type does.
// Signatures are collected as "Name (params) result", where a named parameter
// is "a,b Type" and an unnamed one is the type alone.
//
// Names nested inside a parameter type, as in "fn func(path string) error",
// are left alone; renaming one of those still reads as a change.
func normalizeSignature(signature string) string {
	open := strings.IndexByte(signature, '(')
	if open < 0 {
		return signature
	}
	end := matchParen(signature, open)
	if end < 0 {
		return signature
	}

	params := splitParams(signature[open+1 : end])
	for i, param := range params {
		params[i] = stripParamName(param)
	}
	return signature[:open+1] + strings.Join(params, ", ") + signature[end:]
}

// matchParen returns the index of the parenthesis closing the one at open, or
// -1 when the signature is unbalanced.
func matchParen(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// splitParams splits a parameter list on the ", " separating one parameter from
// the next, leaving the ones nested inside a type such as "func(a, b int)"
// where they are. The names of a single parameter are joined without a space,
// so "a,b int" stays whole.
func splitParams(params string) []string {
	if strings.TrimSpace(params) == "" {
		return nil
	}

	var (
		result []string
		depth  int
		start  int
	)
	for i := 0; i < len(params); i++ {
		switch params[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 && i+1 < len(params) && params[i+1] == ' ' {
				result = append(result, params[start:i])
				start = i + 2
				i++
			}
		}
	}
	return append(result, params[start:])
}

// stripParamName removes the leading name list of a parameter. A parameter
// with no name is its type alone, and is returned as it stands.
func stripParamName(param string) string {
	param = strings.TrimSpace(param)
	space := strings.IndexByte(param, ' ')
	if space < 0 || !isNameList(param[:space]) {
		return param
	}
	return param[space+1:]
}

// isNameList reports whether s is the comma separated name list of a parameter
// rather than the start of a type. Keywords are not identifiers, so a type such
// as "chan int" or "struct { A int }" is not mistaken for a named parameter.
func isNameList(s string) bool {
	if s == "" {
		return false
	}
	for _, name := range strings.Split(s, ",") {
		if !token.IsIdentifier(name) {
			return false
		}
	}
	return true
}

// exportedSymbol reports whether a declaration is part of the API: the name is
// exported, every name it is qualified by is, and the package is importable.
func exportedSymbol(importPath, name string) bool {
	if isInternal(importPath) {
		return false
	}
	for _, part := range strings.Split(name, ".") {
		if !ast.IsExported(part) {
			return false
		}
	}
	return true
}
