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

	names := decl.Names
	if len(names) == 0 {
		if decl.Name == "" {
			return ""
		}
		names = []string{decl.Name}
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
