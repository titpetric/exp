package modules

import (
	"go/ast"
	"strings"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// Documentation computes documentation stats for exported functions in defs.
// Godoc is counted if the comment starts with the function name followed by space and ended with punctuation.
func Documentation(defs model.DefinitionList) DocumentationResponse {
	res := NewDocumentationResponse()
	for _, def := range defs {
		res.Merge(DocumentationForDefinition(def))
	}
	return res
}

func DocumentationForDefinition(def *model.Definition) DocumentationResponse {
	docs := NewDocumentationResponse()
	docs.Packages = 1

	doc := strings.TrimSpace(def.Doc)
	if doc != "" {
		docs.PackageDocs++
	}

	for _, fn := range def.Funcs {
		if fn == nil || !ast.IsExported(fn.Name) {
			continue
		}

		docs.Symbols++

		doc := strings.TrimSpace(fn.Doc)
		if doc != "" {
			docs.SymbolDocs++
			if followsGodoc(fn.Name, doc) {
				docs.SymbolGodocs++
			}
		}
	}

	return docs
}

// followsGodoc returns true if doc starts with name and ends with punctuation.
func followsGodoc(name, doc string) bool {
	doc = strings.TrimSpace(doc)
	if !strings.HasPrefix(doc, name) {
		return false
	}
	if len(doc) == 0 {
		return false
	}

	// check last character of the doc
	last := doc[len(doc)-1]
	switch last {
	case '.', '!', '?':
		return true
	default:
		return false
	}
}
