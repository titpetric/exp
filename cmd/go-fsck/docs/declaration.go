package docs

import (
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// declaration returns what a type, const or var is printed as.
//
// It's the source when the model carries it. A model extracted without
// --include-sources holds the kind, the name and the doc comment but no
// source, and those are written out instead: printing the source alone left
// an empty code block where the symbol should be.
func declaration(decl *model.Declaration) string {
	if source := strings.TrimSpace(decl.Source); source != "" {
		return source
	}

	names := declarationNames(decl)
	if len(names) == 0 {
		return ""
	}

	var out strings.Builder

	if doc := strings.TrimSpace(decl.Doc); doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			if line == "" {
				out.WriteString("//\n")
				continue
			}
			out.WriteString("// " + line + "\n")
		}
	}

	out.WriteString(decl.Kind.String() + " " + strings.Join(names, ", "))
	if decl.Type != "" {
		out.WriteString(" " + decl.Type)
	}

	return out.String()
}

// declarationSummary returns the line a reader picks a declaration out by,
// which is its kind and the names it declares.
func declarationSummary(decl *model.Declaration) string {
	names := declarationNames(decl)
	if len(names) == 0 {
		return decl.Kind.String()
	}

	return decl.Kind.String() + " " + strings.Join(names, ", ")
}

// declarationNames returns the names a declaration carries. A const or var
// block declares more than one.
func declarationNames(decl *model.Declaration) []string {
	if len(decl.Names) > 0 {
		return decl.Names
	}
	if decl.Name != "" {
		return []string{decl.Name}
	}

	return nil
}
