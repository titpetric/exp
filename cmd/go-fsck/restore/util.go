package restore

import (
	"strings"

	"github.com/stoewer/go-strcase"
)

// toFilename is the file a symbol of this name belongs in, which is the snake
// case of the name.
//
// It is the case splint's grouping linter checks a package against, so a
// package restored by this command is one that linter has nothing to say
// about. The two spellings that read wrong in snake case are corrected first:
// OAuth is one word and CoProcess is two.
func toFilename(s string) string {
	s = strings.ReplaceAll(s, "OAuth", "Oauth")
	s = strings.ReplaceAll(s, "CoProcess", "Coprocess")
	s = strcase.SnakeCase(s)
	// hack
	if s == "" {
		return "funcs.go"
	}
	return s + ".go"
}
