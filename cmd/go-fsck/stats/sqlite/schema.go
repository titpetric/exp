package sqlite

import (
	"regexp"
	"strings"

	_ "embed"
)

//go:embed schema.sql
var schema string

// Statements returns a list of statements expanded from the schema.
func Statements() []string {
	result := []string{}

	// remove sql comments from anywhere ([whitespace]--*\n)
	comments := regexp.MustCompile(`\s*--.*`)
	contents := comments.ReplaceAll([]byte(schema), nil)

	// split statements by trailing ; at the end of the line
	stmts := regexp.MustCompilePOSIX(`;$`).Split(string(contents), -1)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}
