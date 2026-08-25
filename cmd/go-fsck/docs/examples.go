package docs

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// collectExamples returns the example functions of each package, keyed by the
// import path of the package they document.
//
// Examples live in a test package, which the reference otherwise leaves out,
// and they are the one symbol worth printing whole: an example is its body.
// They are there only when the model was extracted with --include-tests and
// --include-sources.
func collectExamples(defs []*model.Definition) map[string]model.DeclarationList {
	result := map[string]model.DeclarationList{}

	for _, def := range defs {
		if !def.Package.TestPackage && !strings.HasSuffix(def.Package.Package, "_test") {
			continue
		}

		importPath := strings.TrimSuffix(def.Package.ImportPath, "_test")
		for _, fn := range def.Funcs {
			if !isExample(fn) {
				continue
			}
			result[importPath] = append(result[importPath], fn)
		}
	}

	for _, items := range result {
		items.Sort()
	}

	return result
}

// isExample reports whether a declaration is a godoc example: a function of a
// test file named Example, ExampleF or ExampleF_suffix, taking no arguments
// and returning nothing. The character after "Example" is what tells an
// example from a function that merely starts with the word.
func isExample(fn *model.Declaration) bool {
	if fn.Kind != model.FuncKind || !fn.IsTestScope() {
		return false
	}
	if fn.Receiver != "" || len(fn.Arguments) > 0 || len(fn.Returns) > 0 {
		return false
	}

	suffix, ok := strings.CutPrefix(fn.Name, "Example")
	if !ok {
		return false
	}
	if suffix == "" {
		return true
	}

	return !unicode.IsLower(rune(suffix[0]))
}

// renderExamples returns the examples of one package, each wrapped in a
// <section> named after the function, so a reader arrives at the example
// rather than at the heading above it.
//
// An example with no source behind it is left out: the model was extracted
// without --include-sources, and the signature of an example says nothing.
func renderExamples(items model.DeclarationList) string {
	var out strings.Builder

	for _, fn := range items {
		source := exampleSource(fn)
		if source == "" {
			continue
		}

		fmt.Fprintf(&out, "<section name=%q>\n\n", fn.Name)
		fmt.Fprintf(&out, "### %s\n\n", fn.Name)

		if doc := strings.TrimSpace(fn.Doc); doc != "" {
			fmt.Fprintf(&out, "%s\n\n", fenceCodeBlocks(doc))
		}

		fmt.Fprintf(&out, "```go\n%s\n```\n\n", source)
		fmt.Fprint(&out, "</section>\n\n")
	}

	if out.Len() == 0 {
		return ""
	}

	return "## Examples\n\n" + out.String()
}

// exampleSource returns the example without the doc comment above it, which
// the section prints as prose of its own.
func exampleSource(fn *model.Declaration) string {
	source := strings.TrimSpace(fn.Source)

	for strings.HasPrefix(source, "//") {
		_, rest, ok := strings.Cut(source, "\n")
		if !ok {
			return ""
		}
		source = strings.TrimSpace(rest)
	}

	return source
}
