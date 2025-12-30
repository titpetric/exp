// Package rules provides individual linter implementations for go-fsck.
package rules

import (
	"fmt"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// GodocIssue represents a godoc linting issue.
type GodocIssue struct {
	File        string
	Line        int
	Symbol      string
	Receiver    string
	IssueType   string
	Description string
}

// String formats the godoc issue as a string.
func (g *GodocIssue) String() string {
	loc := fmt.Sprintf("%s:%d", g.File, g.Line)
	symbol := g.Symbol
	if g.Receiver != "" {
		symbol = g.Receiver + "." + symbol
	}
	return fmt.Sprintf("%s %s: %s - %s", loc, symbol, g.IssueType, g.Description)
}

// GodocLinter checks godoc compliance for exported symbols.
type GodocLinter struct {
	issues []*GodocIssue
}

// NewGodocLinter creates a new godoc linter.
func NewGodocLinter() *GodocLinter {
	return &GodocLinter{
		issues: []*GodocIssue{},
	}
}

// Lint checks the declarations for godoc compliance.
func (g *GodocLinter) Lint(defs []*model.Definition) {
	for _, def := range defs {
		isMain := def.Package.Pkg != nil && def.Package.Pkg.Name == "main"
		if isMain || def.Package.TestPackage {
			continue
		}
		g.checkDeclarationList(def, def.Types)
		g.checkDeclarationList(def, def.Funcs)
		g.checkDeclarationList(def, def.Consts)
		g.checkDeclarationList(def, def.Vars)
	}
}

func (g *GodocLinter) checkDeclarationList(def *model.Definition, decls model.DeclarationList) {
	for _, decl := range decls {
		// Only check exported symbols
		if !decl.IsExported() {
			continue
		}

		if decl.IsTestScope() {
			continue
		}

		// Validate godoc format
		g.validateGodoc(def.Package, decl)
	}
}

func (g *GodocLinter) validateGodoc(pkg model.Package, decl *model.Declaration) {
	// Check if godoc exists
	decl.Doc = strings.TrimSpace(decl.Doc)
	doc := decl.Doc
	symbol := decl.Name

	if decl.Doc == "" {
		g.issues = append(g.issues, &GodocIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "missing-godoc",
			Description: "exported symbol lacks godoc comment",
		})
		return
	}

	// Check if comment starts with symbol name
	words := strings.Fields(doc)
	if len(words) == 0 {
		return
	}

	firstWord := words[0]
	if !strings.EqualFold(firstWord, symbol) && !strings.EqualFold(firstWord, decl.Name) {
		g.issues = append(g.issues, &GodocIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "godoc-format",
			Description: fmt.Sprintf("godoc should start with %q, but starts with %q", symbol, firstWord),
		})
		return
	}

	// Check if comment ends with punctuation
	lastChar := doc[len(doc)-1]
	if !hasFinalPunctuation(lastChar) {
		g.issues = append(g.issues, &GodocIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "godoc-format",
			Description: "godoc should end with punctuation (., !, or ?)",
		})
		return
	}

	// Count newlines (hints at overly verbose docs)
	lineCount := strings.Count(doc, "\n")
	if lineCount > 10 {
		g.issues = append(g.issues, &GodocIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "godoc-verbose",
			Description: fmt.Sprintf("godoc is lengthy (%d lines) - may indicate code smell", lineCount+1),
		})
		return
	}
}

func hasFinalPunctuation(ch byte) bool {
	return ch == '.' || ch == '!' || ch == '?' || ch == '`'
}

// Issues returns all godoc issues found.
func (g *GodocLinter) Issues() []*GodocIssue {
	return g.issues
}

// IssueSummary returns statistics about the issues.
func (g *GodocLinter) IssueSummary() map[string]int {
	summary := make(map[string]int)
	for _, issue := range g.issues {
		summary[issue.IssueType]++
	}
	return summary
}
