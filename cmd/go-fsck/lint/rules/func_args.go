package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// FuncArgsIssue represents a function argument ordering issue.
type FuncArgsIssue struct {
	File        string
	Line        int
	Symbol      string
	Receiver    string
	IssueType   string
	Description string
	PackagePath string
}

// String formats the func args issue as a string.
func (f *FuncArgsIssue) String() string {
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

// FuncArgsLinter checks function argument ordering.
type FuncArgsLinter struct {
	issues          []*FuncArgsIssue
	totalSymbols    int
	consideredFuncs int
	passingFuncs    int
	defs            []*model.Definition // Stored for interface type lookup
	argCountStats   map[int]int         // Count of functions by argument count
	argCountValid   map[int]int         // Count of valid functions by argument count
}

// NewFuncArgsLinter creates a new func args linter.
func NewFuncArgsLinter() *FuncArgsLinter {
	return &FuncArgsLinter{
		issues:        []*FuncArgsIssue{},
		argCountStats: make(map[int]int),
		argCountValid: make(map[int]int),
	}
}

// Lint checks function argument ordering in definitions.
func (fa *FuncArgsLinter) Lint(defs []*model.Definition) {
	fa.defs = defs // Store for interface type lookup
	for _, def := range defs {
		fa.checkDeclarationList(def, def.Funcs)
	}
}

func (fa *FuncArgsLinter) checkDeclarationList(def *model.Definition, decls model.DeclarationList) {
	for _, decl := range decls {
		fa.totalSymbols++

		// Only check exported functions
		if !decl.IsExported() {
			continue
		}

		// Skip test scope
		if decl.IsTestScope() {
			continue
		}

		argCount := len(decl.Arguments)
		fa.argCountStats[argCount]++

		// Functions with 0 or 1 argument are always valid
		if argCount < 2 {
			fa.argCountValid[argCount]++
			continue
		}

		fa.consideredFuncs++

		// Check for issues
		if !fa.checkFunctionArgs(def, decl) {
			fa.passingFuncs++
			fa.argCountValid[argCount]++
		}
	}
}

// checkFunctionArgs validates function argument ordering.
// Returns true if issues were found.
func (fa *FuncArgsLinter) checkFunctionArgs(def *model.Definition, decl *model.Declaration) bool {
	args := decl.Arguments

	// Separate variadic and non-variadic arguments (variadics must be last by Go language rules)
	nonVariadic := getNonVariadicArgs(args)
	variadics := getVariadicArgs(args)

	// Check if time.Duration is present - it must be the last non-variadic argument
	// This takes precedence over other ordering rules
	if hasTimeDuration(nonVariadic) && !isTimeDurationLast(nonVariadic) {
		expectedOrder := moveTimeDurationToEnd(nonVariadic)
		if len(variadics) > 0 {
			expectedOrder = append(expectedOrder, variadics...)
		}
		fa.issues = append(fa.issues, &FuncArgsIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "arg-order",
			Description: fmt.Sprintf("expected arguments reorder, %v => %v", args, expectedOrder),
			PackagePath: def.Package.Path,
		})
		return true
	}

	// Check for duplicate types (in non-variadic args)
	if hasDuplicateTypes(nonVariadic) {
		fa.issues = append(fa.issues, &FuncArgsIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "duplicate-type",
			Description: fmt.Sprintf("func %s takes multiple params of same type %s", decl.Name, findDuplicateType(nonVariadic)),
			PackagePath: def.Package.Path,
		})
		return true
	}

	// If time.Duration is present and last, don't apply general ordering rules
	// (time.Duration placement takes precedence)
	if hasTimeDuration(nonVariadic) && isTimeDurationLast(nonVariadic) {
		return false
	}

	// Ambiguous cases (both orderings are valid) - don't report:
	// 1. 2 args with only built-in types
	if len(nonVariadic) == 2 && allBuiltinTypes(nonVariadic) {
		return false
	}
	// 2. (string, any) pattern - first arg is string, exactly 2 args (common for KV/cache setters)
	if len(nonVariadic) == 2 && nonVariadic[0] == "string" {
		return false
	}

	// Check argument order for non-variadic args using stored definitions for interface detection
	expectedOrder := getExpectedOrderWithDefs(nonVariadic, fa.defs)
	if !isCorrectOrder(nonVariadic, expectedOrder) {
		// Append variadics back to expected order
		if len(variadics) > 0 {
			expectedOrder = append(expectedOrder, variadics...)
		}
		fa.issues = append(fa.issues, &FuncArgsIssue{
			File:        decl.File,
			Line:        decl.Line,
			Symbol:      decl.Name,
			Receiver:    decl.Receiver,
			IssueType:   "arg-order",
			Description: fmt.Sprintf("expected arguments reorder, %v => %v", args, expectedOrder),
			PackagePath: def.Package.Path,
		})
		return true
	}

	return false
}

// hasDuplicateTypes checks if there are multiple parameters of the same type.
func hasDuplicateTypes(args []string) bool {
	seen := make(map[string]bool)
	for _, arg := range args {
		if seen[arg] {
			return true
		}
		seen[arg] = true
	}
	return false
}

// findDuplicateType returns the first duplicate type found.
func findDuplicateType(args []string) string {
	seen := make(map[string]bool)
	for _, arg := range args {
		if seen[arg] {
			return arg
		}
		seen[arg] = true
	}
	return ""
}

// isTypeAnInterface checks if a type name is actually an interface by looking it up in the definitions.
func isTypeAnInterface(typeName string, defs []*model.Definition) bool {
	// Handle package.Type format (e.g., "fs.FS", "io.Reader")
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		if len(parts) == 2 {
			pkgName := parts[0]
			typeNameOnly := parts[1]
			// Look for the type in the definitions
			for _, def := range defs {
				// Check if this is the right package (by package name suffix)
				if strings.HasSuffix(def.Package.Path, pkgName) || def.Package.Pkg != nil && def.Package.Pkg.Name == pkgName {
					return isInterfaceInDef(typeNameOnly, def)
				}
			}
		}
		return false
	}

	// For unqualified names, check in current and imported packages
	for _, def := range defs {
		if isInterfaceInDef(typeName, def) {
			return true
		}
	}
	return false
}

// isInterfaceInDef checks if a type name is an interface in a specific definition.
func isInterfaceInDef(typeName string, def *model.Definition) bool {
	for _, typeDecl := range def.Types {
		if typeDecl.Name == typeName && typeDecl.Type == "interface" {
			return true
		}
	}
	return false
}

// getNonVariadicArgs returns only the non-variadic arguments.
func getNonVariadicArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "...") {
			result = append(result, arg)
		}
	}
	return result
}

// getVariadicArgs returns only the variadic arguments.
func getVariadicArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "...") {
			result = append(result, arg)
		}
	}
	return result
}

// hasTimeDuration checks if time.Duration is in the arguments.
func hasTimeDuration(args []string) bool {
	for _, arg := range args {
		if arg == "time.Duration" {
			return true
		}
	}
	return false
}

// isTimeDurationLast checks if time.Duration is the last argument.
func isTimeDurationLast(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[len(args)-1] == "time.Duration"
}

// moveTimeDurationToEnd moves time.Duration to the end of arguments.
func moveTimeDurationToEnd(args []string) []string {
	result := make([]string, 0, len(args))
	var timeDuration string

	for _, arg := range args {
		if arg == "time.Duration" {
			timeDuration = arg
		} else {
			result = append(result, arg)
		}
	}

	if timeDuration != "" {
		result = append(result, timeDuration)
	}

	return result
}

// hasVariadicArgs checks if any argument is variadic (starts with ...).
func hasVariadicArgs(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "...") {
			return true
		}
	}
	return false
}

// allBuiltinTypes checks if all arguments are built-in types.
func allBuiltinTypes(args []string) bool {
	builtins := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"complex64": true, "complex128": true,
	}

	for _, arg := range args {
		// Remove pointers and slices
		clean := strings.TrimPrefix(arg, "*")
		clean = strings.TrimPrefix(clean, "[]")

		if !builtins[clean] && !isMapOrSliceType(arg) {
			return false
		}
	}
	return true
}

// isMapOrSliceType checks if a type is a map or slice type.
func isMapOrSliceType(arg string) bool {
	return strings.HasPrefix(arg, "map[") || strings.HasPrefix(arg, "[]")
}

// isContextType checks if a type is context.Context.
// It allows other context types (matching suffix).
func isContextType(arg string) bool {
	return strings.HasSuffix(arg, "Context")
}

// isInterfaceType checks if a type looks like an interface.
func isInterfaceType(arg string, interfaces map[string]bool) bool {
	clean := strings.TrimPrefix(arg, "*")
	return interfaces[clean]
}

// isChanType checks if a type is a channel type.
func isChanType(arg string) bool {
	clean := strings.TrimPrefix(arg, "*")
	return strings.HasPrefix(clean, "chan")
}

// isStructType checks if a type looks like a struct (has a dot or starts with capital).
func isStructType(arg string) bool {
	clean := strings.TrimPrefix(arg, "*")
	// Structs typically have dots (package.Type) or are capitalized
	return strings.Contains(clean, ".") || (len(clean) > 0 && clean[0] >= 'A' && clean[0] <= 'Z')
}

// getStructOrder returns the sort order for struct type names.
func getStructOrder(arg string) int {
	clean := strings.TrimPrefix(arg, "*")
	lowerClean := strings.ToLower(clean)

	if strings.Contains(lowerClean, "config") {
		return 0
	}
	if strings.Contains(lowerClean, "option") {
		return 1
	}
	if strings.Contains(lowerClean, "flag") {
		return 2
	}
	return 3
}

// getInterfaceOrder returns the sort order for interface type names.
func getInterfaceOrder(arg string) int {
	clean := strings.TrimPrefix(arg, "*")
	lowerClean := strings.ToLower(clean)

	if strings.Contains(lowerClean, "writer") {
		return 0
	}
	if strings.Contains(lowerClean, "reader") {
		return 1
	}
	return 2
}

// getExpectedOrder returns the arguments in expected order.
// This is a simplified version that doesn't have access to the definitions.
// A better approach would be to pass definitions and check actual interface types.
func getExpectedOrder(args []string) []string {
	return getExpectedOrderWithDefs(args, nil)
}

// getExpectedOrderWithDefs returns arguments in expected order, using definition info for interface detection.
func getExpectedOrderWithDefs(args []string, defs []*model.Definition) []string {
	result := make([]string, 0, len(args))

	// Collect by type category
	var contexts []string
	var interfaces []string
	var chans []string
	var structs []string
	var sliceMaps []string
	var builtins []string

	// First pass: identify interface types
	knownInterfaces := make(map[string]bool)
	for _, arg := range args {
		clean := strings.TrimPrefix(arg, "*")
		// Skip slices and maps
		if strings.HasPrefix(clean, "[]") || strings.HasPrefix(clean, "map[") {
			continue
		}

		// If we have definitions, check if this type is actually an interface
		if defs != nil && isTypeAnInterface(clean, defs) {
			knownInterfaces[clean] = true
			continue
		}

		// Fallback to heuristic-based detection
		// Common interface patterns (case-insensitive name matching)
		lowerClean := strings.ToLower(clean)
		if strings.Contains(lowerClean, "handler") || strings.Contains(lowerClean, "writer") ||
			strings.Contains(lowerClean, "reader") || strings.Contains(lowerClean, "fs") ||
			strings.Contains(lowerClean, "closer") || strings.Contains(lowerClean, "listener") ||
			strings.Contains(lowerClean, "router") || strings.Contains(clean, ".") {
			// Any type with a dot (package.Type) is likely an interface
			// Also check for common interface-like names
			knownInterfaces[clean] = true
		}
	}

	// Categorize arguments
	for _, arg := range args {
		if isContextType(arg) {
			contexts = append(contexts, arg)
		} else if isInterfaceType(arg, knownInterfaces) {
			interfaces = append(interfaces, arg)
		} else if isChanType(arg) {
			chans = append(chans, arg)
		} else if isStructType(arg) {
			structs = append(structs, arg)
		} else if isMapOrSliceType(arg) {
			sliceMaps = append(sliceMaps, arg)
		} else {
			builtins = append(builtins, arg)
		}
	}

	// Sort each category
	sortInterfaces(interfaces)
	sortStructs(structs)

	// Append in order
	result = append(result, contexts...)
	result = append(result, interfaces...)
	result = append(result, chans...)
	result = append(result, structs...)
	result = append(result, sliceMaps...)
	result = append(result, builtins...)

	return result
}

// sortInterfaces sorts interface types by preferred order.
func sortInterfaces(interfaces []string) {
	// Simple bubble sort for small slices
	for i := 0; i < len(interfaces)-1; i++ {
		for j := i + 1; j < len(interfaces); j++ {
			if getInterfaceOrder(interfaces[i]) > getInterfaceOrder(interfaces[j]) {
				interfaces[i], interfaces[j] = interfaces[j], interfaces[i]
			}
		}
	}
}

// sortStructs sorts struct types by preferred order.
func sortStructs(structs []string) {
	// Simple bubble sort for small slices
	for i := 0; i < len(structs)-1; i++ {
		for j := i + 1; j < len(structs); j++ {
			if getStructOrder(structs[i]) > getStructOrder(structs[j]) {
				structs[i], structs[j] = structs[j], structs[i]
			}
		}
	}
}

// isCorrectOrder checks if arguments are in the correct order.
func isCorrectOrder(args []string, expected []string) bool {
	if len(args) != len(expected) {
		return false
	}
	for i := range args {
		if args[i] != expected[i] {
			return false
		}
	}
	return true
}

// Issues returns all func args issues found.
func (fa *FuncArgsLinter) Issues() []*FuncArgsIssue {
	return fa.issues
}

// IssueSummary returns statistics about the issues as a map for backward compatibility.
func (fa *FuncArgsLinter) IssueSummary() map[string]interface{} {
	stats := fa.GetStatistics()
	return map[string]interface{}{
		"total_symbols":      stats.TotalSymbols,
		"considered_funcs":   stats.ConsideredFuncs,
		"passing_funcs":      stats.PassingFuncs,
		"reported_issues":    stats.ReportedIssues,
		"arg_order_issues":   stats.ArgOrderIssues,
		"duplicate_issues":   stats.DuplicateIssues,
		"argument_breakdown": stats.ArgumentBreakdown,
	}
}

// GetStatistics returns structured statistics for YAML output.
// Parameter totalSymbols is provided for consistency with other linters but func-args
// tracks its own total via Lint() method.
func (fa *FuncArgsLinter) GetStatistics(totalSymbols ...int) RuleStatistics {
	// Build argument count breakdown
	argBreakdown := make([]ArgCount, 0)
	for i := 0; i <= 10; i++ {
		if count, ok := fa.argCountStats[i]; ok && count > 0 {
			argBreakdown = append(argBreakdown, ArgCount{
				Arguments: i,
				Functions: count,
				Valid:     fa.argCountValid[i],
			})
		}
	}

	return RuleStatistics{
		TotalSymbols:      fa.totalSymbols,
		ConsideredFuncs:   fa.consideredFuncs,
		PassingFuncs:      fa.passingFuncs,
		ReportedIssues:    len(fa.issues),
		ArgOrderIssues:    fa.countIssueType("arg-order"),
		DuplicateIssues:   fa.countIssueType("duplicate-type"),
		ArgumentBreakdown: argBreakdown,
	}
}

// countIssueType counts issues of a specific type.
func (fa *FuncArgsLinter) countIssueType(issueType string) int {
	count := 0
	for _, issue := range fa.issues {
		if issue.IssueType == issueType {
			count++
		}
	}
	return count
}
