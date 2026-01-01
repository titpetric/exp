// Package rules provides individual linter implementations for go-fsck.
package rules

import (
	"fmt"
	"path/filepath"
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
	PackagePath string // Package path for better file path reporting
}

// String formats the godoc issue as a string.
func (g *GodocIssue) String() string {
	file := g.File
	if g.PackagePath != "" && g.PackagePath != "." {
		file = strings.TrimPrefix(g.PackagePath, "."+string(filepath.Separator)) + string(filepath.Separator) + file
	}
	loc := fmt.Sprintf("%s:%d", file, g.Line)
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

// newIssue creates a new GodocIssue with package path information.
func (g *GodocLinter) newIssue(pkg model.Package, decl *model.Declaration, issueType, description string) *GodocIssue {
	return &GodocIssue{
		File:        decl.File,
		Line:        decl.Line,
		Symbol:      decl.Name,
		Receiver:    decl.Receiver,
		IssueType:   issueType,
		Description: description,
		PackagePath: pkg.Path,
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
	if len(decls) == 0 {
		return
	}

	// Group declarations by file and proximity (for const/var blocks)
	type declGroup struct {
		decls []*model.Declaration
	}

	groups := []declGroup{}
	var currentGroup declGroup

	for i, decl := range decls {
		// Only check exported symbols
		if !decl.IsExported() {
			continue
		}

		if decl.IsTestScope() {
			continue
		}

		// Start a new group if this is the first decl or if there's a gap from the previous
		if i == 0 || (len(currentGroup.decls) > 0 && currentGroup.decls[len(currentGroup.decls)-1].File != decl.File) {
			if len(currentGroup.decls) > 0 {
				groups = append(groups, currentGroup)
			}
			currentGroup = declGroup{decls: []*model.Declaration{decl}}
		} else if len(currentGroup.decls) > 0 {
			// Check if it's in the same block (line proximity heuristic)
			lastDecl := currentGroup.decls[len(currentGroup.decls)-1]
			// If lines are close (const/var block), add to same group
			if decl.Line-lastDecl.Line <= 10 {
				currentGroup.decls = append(currentGroup.decls, decl)
			} else {
				// Line gap suggests new block
				groups = append(groups, currentGroup)
				currentGroup = declGroup{decls: []*model.Declaration{decl}}
			}
		} else {
			currentGroup.decls = append(currentGroup.decls, decl)
		}
	}
	if len(currentGroup.decls) > 0 {
		groups = append(groups, currentGroup)
	}

	// Check each group
	for _, group := range groups {
		if len(group.decls) == 0 {
			continue
		}

		// If it's a single declaration, check it normally
		if len(group.decls) == 1 {
			g.validateGodoc(def.Package, group.decls[0])
			continue
		}

		// For a group of declarations: check if the first one has a comment.
		// If it does, consider all undocumented declarations in the group as having inherited documentation.
		firstHasDoc := strings.TrimSpace(group.decls[0].Doc) != ""

		for _, decl := range group.decls {
			if strings.TrimSpace(decl.Doc) == "" && !firstHasDoc {
				// No doc on this declaration and no group comment
				g.issues = append(g.issues, g.newIssue(def.Package, decl, "missing-godoc", "exported symbol lacks godoc comment"))
			} else if strings.TrimSpace(decl.Doc) != "" {
				// Check format for documented declarations
				g.validateGodoc(def.Package, decl)
			}
		}
	}
}

func (g *GodocLinter) validateGodoc(pkg model.Package, decl *model.Declaration) {
	// Check if godoc exists
	decl.Doc = strings.TrimSpace(decl.Doc)
	doc := decl.Doc
	symbol := decl.Name

	if decl.Doc == "" {
		g.issues = append(g.issues, g.newIssue(pkg, decl, "missing-godoc", "exported symbol lacks godoc comment"))
		return
	}

	// For const/var blocks with multiple declarations, skip the symbol name check
	// (the comment applies to the block, not individual symbols)
	if len(decl.Names) > 1 {
		// Just check for punctuation
		lastChar := doc[len(doc)-1]
		if !hasFinalPunctuation(lastChar) {
			g.issues = append(g.issues, g.newIssue(pkg, decl, "godoc-format", "godoc should end with punctuation (., !, or ?)"))
		}
		return
	}

	// Check if comment starts with symbol name
	words := strings.Fields(doc)
	if len(words) == 0 {
		return
	}

	firstWord := words[0]
	if !strings.EqualFold(firstWord, symbol) && !strings.EqualFold(firstWord, decl.Name) {
		g.issues = append(g.issues, g.newIssue(pkg, decl, "godoc-format", fmt.Sprintf("godoc should start with %q, but starts with %q", symbol, firstWord)))
		return
	}

	// Check if comment ends with punctuation
	lastChar := doc[len(doc)-1]
	if !hasFinalPunctuation(lastChar) {
		g.issues = append(g.issues, g.newIssue(pkg, decl, "godoc-format", "godoc should end with punctuation (., !, or ?)"))
		return
	}

	// Count newlines (hints at overly verbose docs)
	lineCount := strings.Count(doc, "\n")
	if lineCount > 10 {
		g.issues = append(g.issues, g.newIssue(pkg, decl, "godoc-verbose", fmt.Sprintf("godoc is lengthy (%d lines) - may indicate code smell", lineCount+1)))
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
