package lint

import (
	"fmt"
	"strings"

	"github.com/titpetric/exp/cmd/schema-gen/model"
	. "github.com/titpetric/exp/cmd/schema-gen/model"
)

func linterFields(cfg *options, pkgInfo *model.PackageInfo) *LintError {
	// Dump out declarations
	errs := NewLintError()
	for _, decl := range pkgInfo.Declarations {
		for _, typeDecl := range decl.Types {
			for _, fieldDecl := range typeDecl.Fields {
				errs.Append(lintField(cfg, decl, typeDecl, fieldDecl)...)
			}
		}
	}
	return errs
}

func lintField(cfg *options, decl *DeclarationInfo, typeDecl *TypeInfo, fieldDecl *FieldInfo) []string {
	rules := cfg.GetRules()
	result := make([]string, 0, len(rules))

	var (
		name  = fieldDecl.Name
		doc   = fieldDecl.Doc
		field = fieldDecl.Path
	)

	if cfg.verbose {
		fmt.Printf("// %s\n", doc)
		fmt.Printf("%s\n\n", field)
	}

	for _, rule := range rules {
		result = append(result, validateRule("field", rule, field, name, doc))
	}
	return result
}

func validateRule(scope, rule, field, name, doc string) string {
	switch {
	case scope == "field" && rule == "require-field-comment":
		fallthrough
	case scope == "struct" && rule == "require-comment":
		if doc == "" {
			return fmt.Sprintf("[%s] No comment on exposed symbol.", field)
		}
	case rule == "require-dot-or-backtick":
		if doc == "" {
			return ""
		}
		if strings.HasSuffix(doc, ".") || strings.HasSuffix(doc, "`") || strings.HasSuffix(doc, "~") {
			return ""
		}

		return fmt.Sprintf("[%s] Symbol comment must end with [.~`] (dot, markdown code block).\nGot:  %s\n", field, doc)
	case rule == "require-prefix":
		if doc == "" {
			return ""
		}
		prefixes := []string{
			fmt.Sprintf("Deprecated: %s ", name),
			fmt.Sprintf("%s ", name),
			fmt.Sprintf("%s: ", name),
		}
		hasPrefix := func(doc string, prefixes []string) bool {
			for _, prefix := range prefixes {
				if strings.HasPrefix(doc, prefix) {
					return true
				}
			}
			return false
		}(doc, prefixes)

		if !hasPrefix {
			return fmt.Sprintf("[%s] Comment must start with symbol name.\nGot:  %s\n", field, doc)
		}
	default:
	}
	return ""
}
