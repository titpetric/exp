package restore

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strings"

	"github.com/titpetric/tools/splint/model"
)

// imported is one import of the source package: the literal as it was written,
// and the name a file reaches it by.
type imported struct {
	literal string
	path    string
	name    string
}

// importsOf reads the import block of one file of the source package.
//
// A literal is the path in quotes, with the name in front of it where the file
// gives one, which is what the model records. The name is what a selector is
// written under, and is the base of the path where the import carries none.
func importsOf(literals []string) []imported {
	out := make([]imported, 0, len(literals))

	for _, literal := range literals {
		name, quoted := "", literal
		if space := strings.LastIndex(literal, " "); space >= 0 {
			name, quoted = literal[:space], literal[space+1:]
		}

		clean := strings.Trim(quoted, `"`)
		if name == "" {
			name = path.Base(clean)
			// A major version is not a package name: example.com/x/v2 is
			// reached as x.
			if major(name) {
				name = path.Base(path.Dir(clean))
			}
		}

		out = append(out, imported{literal: literal, path: clean, name: name})
	}

	return out
}

// major reports the last element of a path that is a major version rather than
// the name a package is reached by.
func major(name string) bool {
	if len(name) < 2 || name[0] != 'v' {
		return false
	}
	for _, r := range name[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// fileImports are the imports one restored file needs: the ones its
// declarations reach, and the ones their source files carried for a side
// effect.
//
// The names a body reaches are read off the syntax rather than out of the
// text, so a package named in a comment or in a string is not imported for it.
// A blank or a dot import is reached by no name at all, so it cannot be read
// off anything: those belong to the file they were written in, and are carried
// by the declarations that came from it.
func fileImports(def *model.Definition, decls model.DeclarationList, body string) ([]string, error) {
	used, err := qualifiers(body)
	if err != nil {
		return nil, err
	}

	var (
		out   []string
		names = map[string]string{}
		seen  = map[string]bool{}
	)

	add := func(one imported) error {
		if known, taken := names[one.name]; taken && known != one.path {
			return fmt.Errorf("the name %q is %s in one file and %s in another, and both are in this one",
				one.name, known, one.path)
		}
		names[one.name] = one.path

		if seen[one.literal] {
			return nil
		}
		seen[one.literal] = true
		out = append(out, one.literal)
		return nil
	}

	for _, decl := range decls {
		for _, one := range importsOf(def.Imports[decl.File]) {
			switch {
			case one.name == "_", one.name == ".":
				// Reached by no name: it is here for what it registers or for
				// what it puts in scope, and both are properties of the file
				// it was written in.
			case !used[one.name]:
				continue
			}
			if err := add(one); err != nil {
				return nil, err
			}
		}
	}

	sort.Strings(out)
	return out, nil
}

// qualifiers are the names a body reaches something through, which is the left
// half of every selector in it.
func qualifiers(body string) (map[string]bool, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), "", body, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	out := map[string]bool{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok := selector.X.(*ast.Ident); ok {
			out[ident.Name] = true
		}
		return true
	})

	return out, nil
}
