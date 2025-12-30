package rules

import (
	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// ImportsLinter checks for import naming collisions.
type ImportsLinter struct {
	issues []error
}

// NewImportsLinter creates a new imports linter.
func NewImportsLinter() *ImportsLinter {
	return &ImportsLinter{
		issues: []error{},
	}
}

// Lint checks for import collisions in definitions.
func (l *ImportsLinter) Lint(defs []*model.Definition) {
	for _, def := range defs {
		_, importCollisions := def.Imports.Map(def.Imports.All())
		l.issues = append(l.issues, importCollisions...)
	}
}

// Issues returns all import collision issues found.
func (l *ImportsLinter) Issues() []error {
	return l.issues
}

// IssueSummary returns statistics about the issues.
func (l *ImportsLinter) IssueSummary() map[string]int {
	return map[string]int{
		"import-collision": len(l.issues),
	}
}
