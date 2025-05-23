package lint

import "github.com/titpetric/exp/cmd/schema-gen/model"

func linterStructs(cfg *options, pkgInfo *model.PackageInfo) *LintError {
	// Dump out declarations
	errs := NewLintError()
	for _, decl := range pkgInfo.Declarations {
		for _, typeDecl := range decl.Types {
			var (
				doc  = decl.Doc
				name = typeDecl.Name
			)
			if typeDecl.Doc != "" {
				doc = typeDecl.Doc
			}

			for _, rule := range cfg.GetRules() {
				errs.Append(validateRule("struct", rule, name, name, doc))
			}
		}
	}
	return errs
}
