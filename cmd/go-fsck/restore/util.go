package restore

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

var builtInTypes = model.BuiltInTypes

var toType = model.ToType

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

func IsConflicting(names []string) error {
	// The problem with unexported functions is that their imports,
	// when merged, would conflict with another function. For example,
	// when using text/template or html/template, math/rand, crypto/rand,
	// or an internal package matching stdlib (internal/crypto).
	conflicting := map[string]bool{
		"html/template":            true,
		"text/template":            true,
		"math/rand":                true,
		"crypto":                   true,
		"crypto/rand":              true,
		"context":                  true,
		"golang.org/x/net/context": true,
	}
	for _, name := range names {
		clean := name
		if strings.Contains(name, " ") {
			clean = strings.Split(name, " ")[1]
		}
		clean = strings.Trim(clean, `"`)
		if ok, _ := conflicting[clean]; ok {
			return fmt.Errorf("Imports conflict over %s", clean)
		}
	}
	return nil
}
