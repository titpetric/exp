package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// FuncReturnsIssue represents a function return ordering issue.
type FuncReturnsIssue struct {
	File        string
	Line        int
	Symbol      string
	Receiver    string
	IssueType   string
	Description string
	PackagePath string
}

// String formats the func returns issue as a string.
func (f *FuncReturnsIssue) String() string {
	file := f.File
	if f.PackagePath != "" && f.PackagePath != "." {
		file = strings.TrimPrefix(f.PackagePath, "."+string(filepath.Separator)) + string(filepath.Separator) + file
	}
	loc := fmt.Sprintf("%s:%d", file, f.Line)
	symbol := f.Symbol
	if f.Receiver != "" {
		symbol = f.Receiver + "." + symbol
	}
	return fmt.Sprintf("%s: %s %s", loc, symbol, f.Description)
}

// FuncReturnsLinter checks function return value ordering.
type FuncReturnsLinter struct {
	issues           []*FuncReturnsIssue
	totalSymbols     int
	consideredFuncs  int
	passingFuncs     int
	returnCountStats map[int]int // Count of functions by return count
	returnCountValid map[int]int // Count of valid functions by return count
}

// NewFuncReturnsLinter creates a new func returns linter.
func NewFuncReturnsLinter() *FuncReturnsLinter {
	return &FuncReturnsLinter{
		issues:           []*FuncReturnsIssue{},
		returnCountStats: make(map[int]int),
		returnCountValid: make(map[int]int),
	}
}

// Lint checks function return value ordering in definitions.
func (fr *FuncReturnsLinter) Lint(defs []*model.Definition) {
	for _, def := range defs {
		fr.checkDeclarationList(def, def.Funcs)
	}
}

func (fr *FuncReturnsLinter) checkDeclarationList(def *model.Definition, decls model.DeclarationList) {
	for _, decl := range decls {
		fr.totalSymbols++

		// Skip test scope
		if decl.IsTestScope() {
			continue
		}

		retCount := len(decl.Returns)
		fr.returnCountStats[retCount]++

		// Functions with 0 or 1 return value are always valid
		if retCount < 2 {
			fr.returnCountValid[retCount]++
			continue
		}

		fr.consideredFuncs++

		if !fr.checkFunctionReturns(def, decl) {
			fr.passingFuncs++
			fr.returnCountValid[retCount]++
		}
	}
}

// checkFunctionReturns validates function return value ordering.
// Returns true if issues were found.
func (fr *FuncReturnsLinter) checkFunctionReturns(def *model.Definition, decl *model.Declaration) bool {
	returns := decl.Returns

	// Check ordering: error and bool should be last
	expected := getExpectedReturnOrder(returns)
	if !isCorrectOrder(returns, expected) {
		fr.issues = append(fr.issues, &FuncReturnsIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "return-order",
			Description: fmt.Sprintf("expected returns reorder, %v => %v", returns, expected),
			PackagePath: def.Package.Path,
		})
		return true
	}

	return false
}

// getExpectedReturnOrder returns the return types in expected order.
// error and bool should be last, with error after bool if both present.
func getExpectedReturnOrder(returns []string) []string {
	var normal []string
	var bools []string
	var errs []string

	for _, ret := range returns {
		switch ret {
		case "error":
			errs = append(errs, ret)
		case "bool":
			bools = append(bools, ret)
		default:
			normal = append(normal, ret)
		}
	}

	result := make([]string, 0, len(returns))
	result = append(result, normal...)
	result = append(result, bools...)
	result = append(result, errs...)
	return result
}

// Issues returns all func returns issues found.
func (fr *FuncReturnsLinter) Issues() []*FuncReturnsIssue {
	return fr.issues
}

// GetStatistics returns structured statistics for YAML output.
func (fr *FuncReturnsLinter) GetStatistics(totalSymbols ...int) RuleStatistics {
	returnBreakdown := make([]ArgCount, 0)
	for i := 0; i <= 10; i++ {
		if count, ok := fr.returnCountStats[i]; ok && count > 0 {
			returnBreakdown = append(returnBreakdown, ArgCount{
				Arguments: i,
				Functions: count,
				Valid:     fr.returnCountValid[i],
			})
		}
	}

	return RuleStatistics{
		TotalSymbols:      fr.totalSymbols,
		ConsideredFuncs:   fr.consideredFuncs,
		PassingFuncs:      fr.passingFuncs,
		ReportedIssues:    len(fr.issues),
		ArgumentBreakdown: returnBreakdown,
	}
}
